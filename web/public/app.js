// Go Scope Visualizer - Interactive Dependency Graph
class GoScopeVisualizer {
    constructor() {
        this.data = null;
        this.svg = null;
        this.simulation = null;
        this.nodes = [];
        this.links = [];
        this.zoom = null;
        this.hiddenFiles = new Set(); // Track hidden files by path
        this.folderTree = null; // Store folder hierarchy

        // Navigation history
        this.history = [];
        this.historyIndex = -1;
        this.isNavigating = false;

        this.config = {
            width: 0,
            height: 0,
            nodeRadius: 20,
            showLabels: true,
            showExternal: true,
            showDocs: true,
            folderDepth: 'all',
            codeBlockWidth: 400,
            codeBlockMaxLines: 999, // Show all code (no truncation)
            codeBlockFontSize: 11,
        };

        this.init();
    }

    init() {
        // Set up file input
        document.getElementById('file-input').addEventListener('change', (e) => this.loadFile(e));

        // Set up controls
        document.getElementById('zoom-in').addEventListener('click', () => this.zoomIn());
        document.getElementById('zoom-out').addEventListener('click', () => this.zoomOut());
        document.getElementById('reset-view').addEventListener('click', () => this.resetView());
        document.getElementById('toggle-labels').addEventListener('click', () => this.toggleLabels());
        document.getElementById('show-external').addEventListener('change', (e) => this.toggleExternal(e.target.checked));
        document.getElementById('show-docs').addEventListener('change', (e) => {
            this.config.showDocs = e.target.checked;
        });
        document.getElementById('folder-depth').addEventListener('change', (e) => this.filterByFolderDepth(e.target.value));
        document.getElementById('close-code').addEventListener('click', () => this.clearCodePanel());
        document.getElementById('nav-back').addEventListener('click', () => this.navigateBack());
        document.getElementById('nav-forward').addEventListener('click', () => this.navigateForward());

        // Folder tree controls
        document.getElementById('expand-all').addEventListener('click', () => this.expandAllFolders());
        document.getElementById('collapse-all').addEventListener('click', () => this.collapseAllFolders());
        document.getElementById('show-all-files').addEventListener('click', () => this.showAllFiles());

        // Set up resizable splitters
        this.initResizer();
        this.initFolderResizer();

        // Initialize empty graph
        this.initializeGraph();
    }

    initResizer() {
        const handle = document.getElementById('resize-handle');
        const codePanel = document.getElementById('code-panel');
        const mainContent = document.querySelector('.main-content');

        let isResizing = false;

        // Restore saved width from localStorage
        const savedWidth = localStorage.getItem('codePanelWidth');
        if (savedWidth) {
            codePanel.style.flexBasis = savedWidth;
            codePanel.style.flexGrow = '0';
            codePanel.style.flexShrink = '0';
        }

        handle.addEventListener('mousedown', (e) => {
            isResizing = true;
            handle.classList.add('resizing');
            document.body.style.cursor = 'col-resize';
            document.body.style.userSelect = 'none';
        });

        document.addEventListener('mousemove', (e) => {
            if (!isResizing) return;

            const containerRect = mainContent.getBoundingClientRect();
            const newWidth = containerRect.right - e.clientX;

            // Min 300px, max 80% of container
            const minWidth = 300;
            const maxWidth = containerRect.width * 0.8;

            if (newWidth >= minWidth && newWidth <= maxWidth) {
                codePanel.style.flexBasis = `${newWidth}px`;
                codePanel.style.flexGrow = '0';
                codePanel.style.flexShrink = '0';
            }
        });

        document.addEventListener('mouseup', () => {
            if (isResizing) {
                isResizing = false;
                handle.classList.remove('resizing');
                document.body.style.cursor = '';
                document.body.style.userSelect = '';

                // Save width to localStorage
                const width = codePanel.style.flexBasis;
                localStorage.setItem('codePanelWidth', width);
            }
        });
    }

    initFolderResizer() {
        const handle = document.getElementById('folder-resize-handle');
        const folderPanel = document.getElementById('folder-panel');
        const mainContent = document.querySelector('.main-content');

        let isResizing = false;

        // Restore saved width from localStorage
        const savedWidth = localStorage.getItem('folderPanelWidth');
        if (savedWidth) {
            folderPanel.style.flexBasis = savedWidth;
            folderPanel.style.flexGrow = '0';
            folderPanel.style.flexShrink = '0';
        }

        handle.addEventListener('mousedown', (e) => {
            isResizing = true;
            handle.classList.add('resizing');
            document.body.style.cursor = 'col-resize';
            document.body.style.userSelect = 'none';
        });

        document.addEventListener('mousemove', (e) => {
            if (!isResizing) return;

            const containerRect = mainContent.getBoundingClientRect();
            const folderRect = folderPanel.getBoundingClientRect();
            const newWidth = e.clientX - folderRect.left;

            // Min 250px, max 500px
            const minWidth = 250;
            const maxWidth = 500;

            if (newWidth >= minWidth && newWidth <= maxWidth) {
                folderPanel.style.flexBasis = `${newWidth}px`;
                folderPanel.style.flexGrow = '0';
                folderPanel.style.flexShrink = '0';
            }
        });

        document.addEventListener('mouseup', () => {
            if (isResizing) {
                isResizing = false;
                handle.classList.remove('resizing');
                document.body.style.cursor = '';
                document.body.style.userSelect = '';

                // Save width to localStorage
                const width = folderPanel.style.flexBasis;
                localStorage.setItem('folderPanelWidth', width);
            }
        });
    }

    initializeGraph() {
        const container = document.getElementById('graph-container');
        this.config.width = container.clientWidth;
        this.config.height = container.clientHeight;

        // Create SVG
        this.svg = d3.select('#graph-container')
            .append('svg')
            .attr('width', '100%')
            .attr('height', '100%')
            .attr('viewBox', [0, 0, this.config.width, this.config.height]);

        // Add zoom behavior
        this.zoom = d3.zoom()
            .scaleExtent([0.1, 4])
            .on('zoom', (event) => {
                this.svg.select('g').attr('transform', event.transform);
            });

        this.svg.call(this.zoom);

        // Create container group
        this.svg.append('g').attr('class', 'graph-group');
    }

    async loadFile(event) {
        const file = event.target.files[0];
        if (!file) return;

        document.getElementById('file-name').textContent = file.name;

        try {
            const text = await file.text();
            this.data = JSON.parse(text);

            // Validate required fields
            if (!this.data.target) {
                throw new Error('Missing "target" field in JSON');
            }
            if (!this.data.nodes) {
                this.data.nodes = [];
            }
            if (!this.data.edges) {
                this.data.edges = [];
            }
            if (!this.data.external) {
                this.data.external = [];
            }

            console.log('Loaded data:', {
                target: this.data.target.name,
                nodes: this.data.nodes.length,
                edges: this.data.edges.length,
                external: this.data.external.length,
                interfaceMappings: (this.data.interfaceMappings || []).length,
                diBindings: (this.data.diBindings || []).length,
                diFramework: this.data.detectedDIFramework || 'none'
            });

            this.renderGraph();
            this.updateStats();
            this.buildFolderTree();
        } catch (error) {
            console.error('Error loading file:', error);
            alert('Error loading JSON file: ' + error.message + '\n\nPlease ensure it\'s a valid go-scope JSON extract.');
        }
    }

    buildFolderTree() {
        if (!this.data || !this.data.nodes) return;

        // Build folder hierarchy from file paths
        const root = { name: 'root', children: {}, files: [], path: '' };

        this.data.nodes.forEach(node => {
            if (!node.file) return;

            const parts = node.file.split('/').filter(p => p);
            let current = root;
            let currentPath = '';

            // Build folder structure
            for (let i = 0; i < parts.length - 1; i++) {
                const part = parts[i];
                currentPath += (currentPath ? '/' : '') + part;

                if (!current.children[part]) {
                    current.children[part] = {
                        name: part,
                        children: {},
                        files: [],
                        path: currentPath
                    };
                }
                current = current.children[part];
            }

            // Add file to current folder
            const fileName = parts[parts.length - 1];
            current.files.push({
                name: fileName,
                path: node.file,
                node: node
            });
        });

        this.folderTree = root;
        this.renderFolderTree();
    }

    renderFolderTree() {
        const container = document.getElementById('folder-tree');
        container.innerHTML = '';

        const renderFolder = (folder, parentElement, depth = 0) => {
            // Render subfolders
            Object.keys(folder.children).sort().forEach(folderName => {
                const subfolder = folder.children[folderName];
                const folderItem = document.createElement('div');
                folderItem.className = 'folder-item';
                folderItem.dataset.path = subfolder.path;

                const folderHeader = document.createElement('div');
                folderHeader.className = 'folder-header';
                folderHeader.style.display = 'flex';
                folderHeader.style.alignItems = 'center';
                folderHeader.style.gap = '0.5rem';
                folderHeader.style.width = '100%';

                // Folder icon (collapsible)
                const icon = document.createElement('span');
                icon.className = 'folder-icon';
                icon.textContent = '📁';
                icon.style.cursor = 'pointer';
                folderHeader.appendChild(icon);

                // Folder name
                const nameSpan = document.createElement('span');
                nameSpan.className = 'folder-name';
                nameSpan.textContent = folderName;
                nameSpan.title = subfolder.path;
                folderHeader.appendChild(nameSpan);

                // Count all files in folder recursively
                const countFiles = (f) => {
                    let count = f.files.length;
                    Object.values(f.children).forEach(child => count += countFiles(child));
                    return count;
                };
                const fileCount = countFiles(subfolder);

                const countSpan = document.createElement('span');
                countSpan.className = 'file-count';
                countSpan.textContent = `(${fileCount})`;
                folderHeader.appendChild(countSpan);

                // Visibility toggle
                const toggle = document.createElement('button');
                toggle.className = 'visibility-toggle';
                toggle.textContent = '👁️';
                toggle.title = 'Hide folder';
                toggle.onclick = (e) => {
                    e.stopPropagation();
                    this.toggleFolderVisibility(subfolder.path);
                };
                folderHeader.appendChild(toggle);

                folderItem.appendChild(folderHeader);

                // Folder children container
                const childrenContainer = document.createElement('div');
                childrenContainer.className = 'folder-children';
                folderItem.appendChild(childrenContainer);

                // Toggle folder collapse
                icon.onclick = () => {
                    folderItem.classList.toggle('collapsed');
                    icon.textContent = folderItem.classList.contains('collapsed') ? '📁' : '📂';
                };

                parentElement.appendChild(folderItem);

                // Render children recursively
                renderFolder(subfolder, childrenContainer, depth + 1);
            });

            // Render files
            folder.files.sort((a, b) => a.name.localeCompare(b.name)).forEach(file => {
                const fileItem = document.createElement('div');
                fileItem.className = 'file-item';
                fileItem.dataset.path = file.path;

                const fileIcon = document.createElement('span');
                fileIcon.className = 'file-icon';
                fileIcon.textContent = '📄';
                fileItem.appendChild(fileIcon);

                const fileName = document.createElement('span');
                fileName.className = 'file-name';
                fileName.textContent = file.name;
                fileName.title = file.path;
                fileItem.appendChild(fileName);

                // Visibility toggle
                const toggle = document.createElement('button');
                toggle.className = 'visibility-toggle';
                toggle.textContent = '👁️';
                toggle.title = 'Hide file';
                toggle.onclick = (e) => {
                    e.stopPropagation();
                    this.toggleFileVisibility(file.path);
                };
                fileItem.appendChild(toggle);

                // Click to show file details
                fileItem.onclick = () => {
                    const node = this.nodes.find(n => n.file === file.path && n.kind === 'file');
                    if (node) {
                        this.showNodeDetails(node);
                        this.highlightNode(node);
                    }
                };

                parentElement.appendChild(fileItem);
            });
        };

        renderFolder(this.folderTree, container);
    }

    toggleFileVisibility(filePath) {
        if (this.hiddenFiles.has(filePath)) {
            this.hiddenFiles.delete(filePath);
        } else {
            this.hiddenFiles.add(filePath);
        }
        this.updateFolderTreeUI();
        this.renderGraph();
    }

    toggleFolderVisibility(folderPath) {
        // Toggle all files in folder and subfolders
        const affectedFiles = this.data.nodes
            .filter(n => n.file && n.file.startsWith(folderPath + '/'))
            .map(n => n.file);

        const allHidden = affectedFiles.every(f => this.hiddenFiles.has(f));

        if (allHidden) {
            // Show all
            affectedFiles.forEach(f => this.hiddenFiles.delete(f));
        } else {
            // Hide all
            affectedFiles.forEach(f => this.hiddenFiles.add(f));
        }

        this.updateFolderTreeUI();
        this.renderGraph();
    }

    updateFolderTreeUI() {
        // Update visibility toggle buttons
        document.querySelectorAll('.file-item').forEach(item => {
            const path = item.dataset.path;
            const toggle = item.querySelector('.visibility-toggle');
            if (this.hiddenFiles.has(path)) {
                item.classList.add('hidden');
                toggle.classList.add('hidden');
                toggle.textContent = '🚫';
                toggle.title = 'Show file';
            } else {
                item.classList.remove('hidden');
                toggle.classList.remove('hidden');
                toggle.textContent = '👁️';
                toggle.title = 'Hide file';
            }
        });

        document.querySelectorAll('.folder-item').forEach(item => {
            const path = item.dataset.path;
            const toggle = item.querySelector('.visibility-toggle');

            // Check if all files in folder are hidden
            const affectedFiles = this.data.nodes
                .filter(n => n.file && n.file.startsWith(path + '/'))
                .map(n => n.file);

            const allHidden = affectedFiles.length > 0 && affectedFiles.every(f => this.hiddenFiles.has(f));

            if (allHidden) {
                item.classList.add('hidden');
                toggle.classList.add('hidden');
                toggle.textContent = '🚫';
                toggle.title = 'Show folder';
            } else {
                item.classList.remove('hidden');
                toggle.classList.remove('hidden');
                toggle.textContent = '👁️';
                toggle.title = 'Hide folder';
            }
        });
    }

    expandAllFolders() {
        document.querySelectorAll('.folder-item').forEach(item => {
            item.classList.remove('collapsed');
            const icon = item.querySelector('.folder-icon');
            if (icon) icon.textContent = '📂';
        });
    }

    collapseAllFolders() {
        document.querySelectorAll('.folder-item').forEach(item => {
            item.classList.add('collapsed');
            const icon = item.querySelector('.folder-icon');
            if (icon) icon.textContent = '📁';
        });
    }

    showAllFiles() {
        this.hiddenFiles.clear();
        this.updateFolderTreeUI();
        this.renderGraph();
    }

    renderGraph() {
        if (!this.data) return;

        // Prepare nodes and links
        this.prepareData();

        // Clear existing graph
        this.svg.select('g').selectAll('*').remove();

        const g = this.svg.select('g');

        // Create force simulation with larger spacing for code blocks
        this.simulation = d3.forceSimulation(this.nodes)
            .force('link', d3.forceLink(this.links)
                .id(d => d.id)
                .distance(d => 300)) // Increased distance for code blocks
            .force('charge', d3.forceManyBody().strength(-1000)) // Stronger repulsion
            .force('center', d3.forceCenter(this.config.width / 2, this.config.height / 2))
            .force('collision', d3.forceCollide()
                .radius(d => Math.max(d.width || 400, d.height || 200) / 2 + 50)) // Account for rectangle size
            .force('x', d3.forceX(this.config.width / 2).strength(0.02))
            .force('y', d3.forceY(this.config.height / 2).strength(0.02));

        // Create links
        const link = g.append('g')
            .attr('class', 'links')
            .selectAll('line')
            .data(this.links)
            .join('line')
            .attr('class', 'link')
            .attr('stroke-width', d => 1 + d.depth * 0.5);

        // Create nodes
        const node = g.append('g')
            .attr('class', 'nodes')
            .selectAll('g')
            .data(this.nodes)
            .join('g')
            .attr('class', d => {
                if (d.isTarget) return 'node target';
                if (d.external) return 'node external';
                if (d.kind === 'interface') return 'node interface';
                if (d.kind === 'struct' && this.isImplementation(d.name)) return 'node implementation';
                if (d.kind === 'func' && d.name.startsWith('New')) return 'node constructor';
                return 'node internal';
            })
            .call(this.drag(this.simulation))
            .on('click', (event, d) => this.showNodeDetails(d));

        // Add code block foreignObjects
        const foreignObject = node.append('foreignObject')
            .attr('width', this.config.codeBlockWidth)
            .attr('height', d => {
                const codeInfo = this.prepareCodeForNode(d);
                d.codeInfo = codeInfo; // Store for later use
                const lineHeight = this.config.codeBlockFontSize * 1.5;
                const headerHeight = 30;
                const footerHeight = codeInfo.totalLines > codeInfo.lines ? 20 : 0;
                return headerHeight + (codeInfo.lines * lineHeight) + footerHeight + 20;
            })
            .attr('x', d => -this.config.codeBlockWidth / 2)
            .attr('y', d => -d.codeInfo.lines * this.config.codeBlockFontSize * 1.5 / 2 - 15);

        // Add HTML content
        foreignObject.append('xhtml:div')
            .attr('class', d => {
                let classes = 'code-node';
                if (d.isTarget) classes += ' target';
                if (d.external) classes += ' external';
                return classes;
            })
            .html(d => d.codeInfo.html);

        // Store dimensions for collision detection
        node.each(d => {
            d.width = this.config.codeBlockWidth;
            d.height = d.codeInfo.lines * this.config.codeBlockFontSize * 1.5 + 50;
        });

        // Update collision force with actual node dimensions
        this.simulation.force('collision', d3.forceCollide()
            .radius(d => {
                // Use diagonal of rectangle + padding for collision
                const width = d.width || 400;
                const height = d.height || 200;
                const diagonal = Math.sqrt(width * width + height * height) / 2;
                return diagonal + 30; // Add 30px padding
            })
            .strength(0.8) // Strong collision avoidance
            .iterations(3)); // Multiple iterations for better separation

        // Restart simulation to apply new collision
        this.simulation.alpha(1).restart();

        // Apply Prism syntax highlighting to all code blocks
        setTimeout(() => {
            if (typeof Prism !== 'undefined') {
                document.querySelectorAll('.code-node-code code').forEach(block => {
                    Prism.highlightElement(block);
                });
            }
        }, 100);

        // Update positions on simulation tick
        this.simulation.on('tick', () => {
            link
                .attr('x1', d => d.source.x)
                .attr('y1', d => d.source.y)
                .attr('x2', d => d.target.x)
                .attr('y2', d => d.target.y);

            node.attr('transform', d => `translate(${d.x},${d.y})`);
        });

        // Highlight target initially
        if (this.data.target) {
            this.showNodeDetails(this.nodes.find(n => n.isTarget));
        }
    }

    prepareData() {
        // Store original symbol data for reference
        this.symbols = [this.data.target, ...this.data.nodes];

        // Group symbols by file to create file nodes
        const fileMap = new Map();

        this.symbols.forEach(symbol => {
            if (!symbol.file) return; // Skip symbols without file info

            if (!fileMap.has(symbol.file)) {
                fileMap.set(symbol.file, {
                    id: symbol.file,
                    name: symbol.file.split('/').pop(), // Just the filename
                    fullPath: symbol.file,
                    file: symbol.file,
                    kind: 'file',
                    package: symbol.package,
                    symbols: [],
                    external: symbol.external,
                    isTarget: symbol.isTarget || false,
                    depth: symbol.depth || 0,
                    exported: symbol.exported
                });
            }

            fileMap.get(symbol.file).symbols.push(symbol);

            // Mark file as target if it contains the target symbol
            if (symbol.isTarget) {
                fileMap.get(symbol.file).isTarget = true;
            }
        });

        // Convert map to array
        this.nodes = Array.from(fileMap.values());

        // Filter hidden files
        this.nodes = this.nodes.filter(n => !this.hiddenFiles.has(n.file));

        // Filter external nodes if needed
        if (!this.config.showExternal) {
            this.nodes = this.nodes.filter(n => !n.external);
        }

        // Filter by folder depth
        this.nodes = this.nodes.filter(n => this.shouldShowNodeByDepth(n));

        // Create file-to-file edges from symbol dependencies
        const fileEdges = new Map();

        this.data.edges.forEach(edge => {
            // Find source and target symbols
            const sourceSymbol = this.symbols.find(s => s.name === edge.from);
            const targetSymbol = this.symbols.find(s => s.name === edge.to);

            if (!sourceSymbol || !targetSymbol) return;
            if (!sourceSymbol.file || !targetSymbol.file) return;
            if (sourceSymbol.file === targetSymbol.file) return; // Skip same-file edges

            const edgeKey = `${sourceSymbol.file}->${targetSymbol.file}`;

            if (!fileEdges.has(edgeKey)) {
                fileEdges.set(edgeKey, {
                    source: sourceSymbol.file,
                    target: targetSymbol.file,
                    count: 0,
                    depth: 1,
                    symbols: []
                });
            }

            const fileEdge = fileEdges.get(edgeKey);
            fileEdge.count++;
            fileEdge.symbols.push({ from: edge.from, to: edge.to });
        });

        // Convert to array
        this.links = Array.from(fileEdges.values());

        // Filter links to only include visible nodes
        const nodeIds = new Set(this.nodes.map(n => n.id));
        this.links = this.links.filter(l =>
            nodeIds.has(l.source) && nodeIds.has(l.target)
        );

        console.log('File-based graph:', {
            files: this.nodes.length,
            links: this.links.length,
            targetFile: this.data.target?.file
        });

        // Debug: Show distance for each file
        if (this.nodes.length > 0 && this.data.target?.file) {
            const targetFile = this.data.target.file;
            const distances = this.nodes.map(n => ({
                file: n.name,
                distance: this.calculateFolderDistance(targetFile, n.file)
            }));
            console.log('All file distances:', distances);

            // Count files at each distance
            const distCounts = {};
            distances.forEach(d => {
                distCounts[d.distance] = (distCounts[d.distance] || 0) + 1;
            });
            console.log('Files per distance:', distCounts);
        }
    }

    calculateFolderDistance(targetFile, nodeFile) {
        if (!targetFile || !nodeFile) return 999;

        const targetParts = targetFile.split('/').filter(p => p);
        const nodeParts = nodeFile.split('/').filter(p => p);

        const targetFolder = targetParts.slice(0, -1).join('/');
        const nodeFolder = nodeParts.slice(0, -1).join('/');

        if (targetFolder === nodeFolder) return 0;

        let commonLength = 0;
        for (let i = 0; i < Math.min(targetParts.length - 1, nodeParts.length - 1); i++) {
            if (targetParts[i] === nodeParts[i]) {
                commonLength = i + 1;
            } else {
                break;
            }
        }

        const targetDepth = targetParts.length - 1;
        const nodeDepth = nodeParts.length - 1;
        const targetSteps = targetDepth - commonLength;
        const nodeSteps = nodeDepth - commonLength;

        return targetSteps + nodeSteps;
    }

    // Prepare code for display in graph nodes
    prepareCodeForNode(node) {
        if (!node.symbols || node.symbols.length === 0) {
            return { html: '<div class="empty-code">No code available</div>', lines: 1 };
        }

        // Combine all symbol code from this file
        let allCode = '';
        let symbolMap = []; // Track which lines belong to which symbol

        node.symbols.forEach((symbol, idx) => {
            if (symbol.code) {
                const lines = symbol.code.split('\n');
                const startLine = allCode.split('\n').length;

                lines.forEach((line, lineIdx) => {
                    symbolMap.push({
                        symbolIdx: idx,
                        symbolName: symbol.name,
                        lineInSymbol: lineIdx,
                        globalLine: startLine + lineIdx
                    });
                });

                allCode += symbol.code + '\n\n';
            }
        });

        // Truncate if too long
        const lines = allCode.split('\n');
        const maxLines = this.config.codeBlockMaxLines;
        let truncated = false;
        let displayLines = lines;

        if (lines.length > maxLines) {
            displayLines = lines.slice(0, maxLines);
            truncated = true;
        }

        // Escape HTML
        const escaped = displayLines.map(l => this.escapeHtml(l)).join('\n');

        const html = `<div class="code-node-content">
            <div class="code-node-header">${node.name}</div>
            <pre class="code-node-code language-go"><code class="language-go">${escaped}</code></pre>
            ${truncated ? '<div class="code-node-truncated">... +' + (lines.length - maxLines) + ' more lines</div>' : ''}
        </div>`;

        return {
            html,
            lines: displayLines.length,
            symbolMap,
            totalLines: lines.length
        };
    }

    showNodeDetails(node, fromHistory = false) {
        // Add to history if not navigating via back/forward
        if (!fromHistory && !this.isNavigating) {
            // Remove any forward history if we're not at the end
            if (this.historyIndex < this.history.length - 1) {
                this.history = this.history.slice(0, this.historyIndex + 1);
            }
            this.history.push(node);
            this.historyIndex = this.history.length - 1;
            this.updateNavigationButtons();
        }

        const codeContent = document.getElementById('code-content');
        const codeTitle = document.getElementById('code-title');

        // Build details HTML
        let html = '<div class="node-details">';

        // For file nodes, show file info
        if (node.kind === 'file') {
            html += `<div class="detail-row">
                <span class="detail-label">File:</span>
                <span class="detail-value"><strong>${node.name}</strong></span>
            </div>`;

            html += `<div class="detail-row">
                <span class="detail-label">Path:</span>
                <span class="detail-value">${node.fullPath}</span>
            </div>`;

            if (node.package) {
                html += `<div class="detail-row">
                    <span class="detail-label">Package:</span>
                    <span class="detail-value">${node.package}</span>
                </div>`;
            }

            html += `<div class="detail-row">
                <span class="detail-label">Symbols:</span>
                <span class="detail-value">${node.symbols.length} symbols</span>
            </div>`;
        } else {
            // Legacy: for individual symbol nodes
            html += `<div class="detail-row">
                <span class="detail-label">Name:</span>
                <span class="detail-value"><strong>${node.name}</strong></span>
            </div>`;

            html += `<div class="detail-row">
                <span class="detail-label">Kind:</span>
                <span class="detail-value">${node.kind}</span>
            </div>`;

            if (node.package) {
                html += `<div class="detail-row">
                    <span class="detail-label">Package:</span>
                    <span class="detail-value">${node.package}</span>
                </div>`;
            }

            if (node.file) {
                const fileName = node.file.split('/').pop();
                const fileLink = this.createFileLink(node.file, node.line);
                html += `<div class="detail-row">
                    <span class="detail-label">Source:</span>
                    <span class="detail-value">${fileLink}</span>
                </div>`;
            }
        }

        html += `<div class="detail-row">
            <span class="detail-label">Exported:</span>
            <span class="detail-value">${node.exported ? '✓ Yes' : '✗ No'}</span>
        </div>`;

        html += `<div class="detail-row">
            <span class="detail-label">Depth:</span>
            <span class="detail-value">${node.depth}</span>
        </div>`;

        if (node.external) {
            html += `<div class="detail-row">
                <span class="detail-label">External:</span>
                <span class="detail-value">✓ External Package</span>
            </div>`;
        }

        html += '</div>';

        // For file nodes, show all symbols in the file
        if (node.kind === 'file' && node.symbols && node.symbols.length > 0) {
            html += '<h3>Symbols in File</h3>';

            // Sort symbols by line number
            const sortedSymbols = [...node.symbols].sort((a, b) => (a.line || 0) - (b.line || 0));

            sortedSymbols.forEach(symbol => {
                if (symbol.doc && this.config.showDocs) {
                    html += `<div class="node-doc">${this.escapeHtml(symbol.doc)}</div>`;
                }

                if (symbol.code) {
                    html += `<h4>${symbol.name} (${symbol.kind})</h4>`;
                    const highlightedCode = this.highlightCode(symbol.code, symbol.name, symbol.kind);
                    html += `<div class="code-block"><pre class="language-go"><code class="language-go">${highlightedCode}</code></pre></div>`;
                }
            });
        } else {
            // Legacy: for individual symbol nodes
            // Add documentation if available
            if (node.doc && this.config.showDocs) {
                html += `<div class="node-doc">${this.escapeHtml(node.doc)}</div>`;
            }

            // Add code if available
            if (node.code) {
                html += '<h3>Code</h3>';
                console.log('Processing code for:', node.name, 'kind:', node.kind);
                const highlightedCode = this.highlightCode(node.code, node.name, node.kind);
                console.log('Highlighted code length:', highlightedCode.length);
                html += `<div class="code-block"><pre class="language-go"><code class="language-go">${highlightedCode}</code></pre></div>`;
            } else if (node.external) {
                html += '<div class="empty-state">External symbol - code not available</div>';
            } else {
                console.warn('No code available for node:', node.name);
            }
        }

        codeTitle.textContent = node.kind === 'file' ? node.name : `${node.name} (${node.kind})`;
        codeContent.innerHTML = html;

        // Check if code-link spans exist before Prism
        const linksBeforePrism = codeContent.querySelectorAll('.code-link');
        console.log('Code links before Prism:', linksBeforePrism.length);

        // Apply Prism highlighting after inserting HTML
        if (typeof Prism !== 'undefined') {
            Prism.highlightAllUnder(codeContent);
        }

        // Check if code-link spans exist after Prism
        const linksAfterPrism = codeContent.querySelectorAll('.code-link');
        console.log('Code links after Prism:', linksAfterPrism.length);

        // Add click handlers for code links
        this.attachCodeLinkHandlers();

        // Update breadcrumb
        this.updateBreadcrumb();
    }

    updateBreadcrumb() {
        const breadcrumb = document.getElementById('breadcrumb');
        if (this.history.length === 0) {
            breadcrumb.innerHTML = '';
            return;
        }

        // Show last 5 items
        const start = Math.max(0, this.historyIndex - 4);
        const items = this.history.slice(start, this.historyIndex + 1);

        breadcrumb.innerHTML = items.map((node, index) => {
            const actualIndex = start + index;
            const isActive = actualIndex === this.historyIndex;
            const className = isActive ? 'breadcrumb-item active' : 'breadcrumb-item';
            const onclick = isActive ? '' : `onclick="window.goScopeViz.navigateToHistory(${actualIndex})"`;

            return `
                ${index > 0 ? '<span class="breadcrumb-separator">›</span>' : ''}
                <span class="${className}" ${onclick}>${node.name}</span>
            `;
        }).join('');
    }

    navigateToHistory(index) {
        if (index < 0 || index >= this.history.length) return;

        this.historyIndex = index;
        this.isNavigating = true;
        this.showNodeDetails(this.history[index], true);
        this.highlightNode(this.history[index]);
        this.isNavigating = false;
        this.updateNavigationButtons();
    }

    navigateBack() {
        if (this.historyIndex > 0) {
            this.historyIndex--;
            this.isNavigating = true;
            const node = this.history[this.historyIndex];
            this.showNodeDetails(node, true);
            this.highlightNode(node);
            this.isNavigating = false;
            this.updateNavigationButtons();
        }
    }

    navigateForward() {
        if (this.historyIndex < this.history.length - 1) {
            this.historyIndex++;
            this.isNavigating = true;
            const node = this.history[this.historyIndex];
            this.showNodeDetails(node, true);
            this.highlightNode(node);
            this.isNavigating = false;
            this.updateNavigationButtons();
        }
    }

    updateNavigationButtons() {
        const backBtn = document.getElementById('nav-back');
        const forwardBtn = document.getElementById('nav-forward');

        backBtn.disabled = this.historyIndex <= 0;
        forwardBtn.disabled = this.historyIndex >= this.history.length - 1;
    }

    attachCodeLinkHandlers() {
        const codeLinks = document.querySelectorAll('.code-link');
        console.log('Found code links:', codeLinks.length);

        codeLinks.forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                e.stopPropagation();
                const nodeId = link.getAttribute('data-node-id');
                const nodeName = link.getAttribute('data-node-name');

                console.log('Code link clicked:', nodeName);

                // Find the node
                const node = this.nodes.find(n => n.id === nodeId || n.name === nodeName);
                if (node) {
                    console.log('Navigating to node:', node.name);
                    // Show the node details
                    this.showNodeDetails(node);

                    // Highlight the node in the graph
                    this.highlightNode(node);
                } else {
                    console.error('Node not found:', nodeName, nodeId);
                }
            });
        });
    }

    highlightNode(node) {
        // Remove previous highlights
        this.svg.selectAll('.node').classed('highlighted', false);

        // Add highlight to the target node
        this.svg.selectAll('.node')
            .filter(d => d.id === node.id)
            .classed('highlighted', true)
            .raise(); // Bring to front

        // Optionally center on the node
        if (node.x && node.y) {
            const transform = d3.zoomIdentity
                .translate(this.config.width / 2, this.config.height / 2)
                .scale(1.5)
                .translate(-node.x, -node.y);

            this.svg.transition()
                .duration(500)
                .call(this.zoom.transform, transform);
        }
    }

    createFileLink(filePath, line) {
        const fileName = filePath.split('/').pop();
        const displayText = `${fileName}:${line}`;

        // Create an ID for this specific copy button
        const buttonId = `copy-${Math.random().toString(36).substr(2, 9)}`;

        // VS Code URI for opening files
        const vscodeUri = `vscode://file${filePath}:${line}`;

        // Create clickable file link and copy button
        const fileLink = `<a href="${vscodeUri}" class="file-link-clickable" title="${filePath}\nClick to open in VS Code">${displayText}</a>`;
        const copyButton = `<button id="${buttonId}" class="copy-btn" data-path="${filePath}:${line}" title="Copy file path">📋</button>`;

        // Set up the click handler after DOM insertion
        setTimeout(() => {
            const btn = document.getElementById(buttonId);
            if (btn) {
                btn.addEventListener('click', (e) => {
                    e.stopPropagation();
                    const path = btn.getAttribute('data-path');
                    navigator.clipboard.writeText(path).then(() => {
                        btn.textContent = '✓';
                        btn.style.color = 'green';
                        setTimeout(() => {
                            btn.textContent = '📋';
                            btn.style.color = '';
                        }, 1000);
                    });
                });
            }
        }, 0);

        return `${fileLink} ${copyButton}`;
    }

    highlightCode(code, symbolName, kind) {
        // DISABLED: app-simple.js now handles linking AFTER Prism
        // let result = this.linkifyIdentifiers(code);

        // Escape HTML
        let result = this.escapeHtml(code);

        // For functions, highlight the function name
        if (kind === 'func' || kind === 'method') {
            const funcPattern = new RegExp(`\\b(func\\s+(?:\\([^)]*\\)\\s+)?)(${this.escapeRegex(symbolName)})(\\s*\\()`, 'g');
            result = result.replace(funcPattern, '$1<mark class="highlight-symbol">$2</mark>$3');
        }

        // For types, highlight the type name
        if (kind === 'struct' || kind === 'interface' || kind === 'type') {
            const typePattern = new RegExp(`\\b(type\\s+)(${this.escapeRegex(symbolName)})(\\s+)`, 'g');
            result = result.replace(typePattern, '$1<mark class="highlight-symbol">$2</mark>$3');
        }

        return result;
    }

    linkifyIdentifiers(code) {
        if (!this.data || !this.nodes) {
            console.log('No data or nodes available for linkifying');
            return code;
        }

        // Build a map of symbol names to nodes
        const symbolMap = new Map();
        this.nodes.forEach(node => {
            if (node.name && !node.external) {
                symbolMap.set(node.name, node);
            }
        });

        console.log('Symbol map size:', symbolMap.size);
        console.log('Available symbols:', Array.from(symbolMap.keys()).slice(0, 20).join(', '));

        // Sort by length (longest first) to avoid partial matches
        const symbols = Array.from(symbolMap.keys()).sort((a, b) => b.length - a.length);

        let linkCount = 0;

        // Replace each symbol with a clickable link
        symbols.forEach(symbolName => {
            const node = symbolMap.get(symbolName);
            const escapedName = this.escapeRegex(symbolName);

            // Simple word boundary match - this runs BEFORE HTML escaping
            const pattern = new RegExp(`\\b(${escapedName})\\b`, 'g');
            const replacement = `\x00LINK_START\x00${node.id}\x00${node.name}\x00$1\x00LINK_END\x00`;

            const before = code;
            code = code.replace(pattern, replacement);
            if (code !== before) {
                linkCount += (code.match(pattern) || []).length;
            }
        });

        console.log('Created', linkCount, 'potential links');
        return code;
    }

    escapeHtml(text) {
        // Simple HTML escape - app-simple.js handles linking
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    escapeRegex(str) {
        return str.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    }

    clearCodePanel() {
        document.getElementById('code-content').innerHTML =
            '<div class="empty-state">👈 Click on a node in the graph to view its code and details</div>';
        document.getElementById('code-title').textContent = 'Select a node to view code';
    }

    updateStats() {
        if (!this.data) return;

        document.getElementById('stat-nodes').textContent = this.nodes.length;
        document.getElementById('stat-edges').textContent = this.data.edges.length;
        document.getElementById('stat-depth').textContent = this.data.totalLayers;
        document.getElementById('stat-external').textContent = this.data.external?.length || 0;

        // Show Phase 3 stats if available
        if (this.data.interfaceMappings && this.data.interfaceMappings.length > 0) {
            console.log(`📐 Arch: ${this.data.interfaceMappings.length} interfaces, DI: ${this.data.detectedDIFramework}`);
        }
    }

    // Check if a node is an implementation in interface mappings
    isImplementation(nodeName) {
        if (!this.data.interfaceMappings) return false;

        for (const mapping of this.data.interfaceMappings) {
            for (const impl of mapping.implementations) {
                if (impl.name === nodeName) return true;
            }
        }
        return false;
    }

    // Zoom controls
    zoomIn() {
        this.svg.transition().call(this.zoom.scaleBy, 1.3);
    }

    zoomOut() {
        this.svg.transition().call(this.zoom.scaleBy, 0.7);
    }

    resetView() {
        this.svg.transition().call(
            this.zoom.transform,
            d3.zoomIdentity.translate(0, 0).scale(1)
        );
    }

    toggleLabels() {
        this.config.showLabels = !this.config.showLabels;
        this.svg.selectAll('.node text')
            .style('display', this.config.showLabels ? 'block' : 'none');
    }

    // Folder depth filtering
    filterByFolderDepth(mode) {
        this.config.folderDepth = mode;
        if (this.data) {
            this.renderGraph();
            this.updateStats();
        }
    }

    calculateFolderDepth(targetFile, nodeFile) {
        if (!targetFile || !nodeFile) return 0;

        const targetParts = targetFile.split('/').filter(p => p);
        const nodeParts = nodeFile.split('/').filter(p => p);

        // Get folder paths (exclude filename)
        const targetFolder = targetParts.slice(0, -1).join('/');
        const nodeFolder = nodeParts.slice(0, -1).join('/');

        // If exact same folder, depth is 0
        if (targetFolder === nodeFolder) {
            return 0;
        }

        // Find common ancestor
        let commonLength = 0;
        for (let i = 0; i < Math.min(targetParts.length - 1, nodeParts.length - 1); i++) {
            if (targetParts[i] === nodeParts[i]) {
                commonLength = i + 1;
            } else {
                break;
            }
        }

        // Calculate depth from common ancestor
        const targetDepth = targetParts.length - 1; // -1 to exclude filename
        const nodeDepth = nodeParts.length - 1;

        // Steps from common ancestor to each folder
        const targetSteps = targetDepth - commonLength;
        const nodeSteps = nodeDepth - commonLength;

        // If target is deeper, node is "up" (negative)
        // If node is deeper, node is "down" (positive)
        if (targetSteps === 0 && nodeSteps > 0) {
            // Node is below target folder
            return nodeSteps;
        } else if (nodeSteps === 0 && targetSteps > 0) {
            // Node is above target folder
            return -targetSteps;
        } else {
            // Different branches - return total distance
            return targetSteps + nodeSteps;
        }
    }

    shouldShowNodeByDepth(node) {
        if (this.config.folderDepth === 'all') return true;
        if (!this.data || !this.data.target) return true;

        const targetFile = this.data.target.file;
        const nodeFile = node.file;

        if (!targetFile || !nodeFile) return true;

        const distance = this.calculateFolderDistance(targetFile, nodeFile);
        const maxDistance = parseInt(this.config.folderDepth);

        return distance <= maxDistance;
    }

    toggleExternal(show) {
        this.config.showExternal = show;
        if (this.data) {
            this.renderGraph();
            this.updateStats();
        }
    }

    // Drag behavior
    drag(simulation) {
        function dragstarted(event, d) {
            if (!event.active) simulation.alphaTarget(0.3).restart();
            d.fx = d.x;
            d.fy = d.y;
        }

        function dragged(event, d) {
            d.fx = event.x;
            d.fy = event.y;
        }

        function dragended(event, d) {
            if (!event.active) simulation.alphaTarget(0);
            d.fx = null;
            d.fy = null;
        }

        return d3.drag()
            .on('start', dragstarted)
            .on('drag', dragged)
            .on('end', dragended);
    }

    // Utility functions
    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
}

// Initialize visualizer when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    const viz = new GoScopeVisualizer();
    window.goScopeViz = viz;
    document.goScopeViz = viz;
});
