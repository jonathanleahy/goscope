// Go Scope Visualizer - Interactive Dependency Graph
class GoScopeVisualizer {
    constructor() {
        this.data = null;
        this.svg = null;
        this.simulation = null;
        this.nodes = [];
        this.links = [];
        this.zoom = null;

        this.config = {
            width: 0,
            height: 0,
            nodeRadius: 20,
            showLabels: true,
            showExternal: true,
            showDocs: true,
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
        document.getElementById('close-code').addEventListener('click', () => this.clearCodePanel());

        // Initialize empty graph
        this.initializeGraph();
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
            this.renderGraph();
            this.updateStats();
        } catch (error) {
            console.error('Error loading file:', error);
            alert('Error loading JSON file. Please ensure it\'s a valid go-scope JSON extract.');
        }
    }

    renderGraph() {
        if (!this.data) return;

        // Prepare nodes and links
        this.prepareData();

        // Clear existing graph
        this.svg.select('g').selectAll('*').remove();

        const g = this.svg.select('g');

        // Create force simulation
        this.simulation = d3.forceSimulation(this.nodes)
            .force('link', d3.forceLink(this.links)
                .id(d => d.id)
                .distance(d => 100 + (d.depth * 50)))
            .force('charge', d3.forceManyBody().strength(-300))
            .force('center', d3.forceCenter(this.config.width / 2, this.config.height / 2))
            .force('collision', d3.forceCollide().radius(this.config.nodeRadius + 10))
            .force('x', d3.forceX(this.config.width / 2).strength(0.05))
            .force('y', d3.forceY(this.config.height / 2).strength(0.05));

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
                return 'node internal';
            })
            .call(this.drag(this.simulation))
            .on('click', (event, d) => this.showNodeDetails(d));

        // Add circles
        node.append('circle')
            .attr('r', d => d.isTarget ? this.config.nodeRadius * 1.5 : this.config.nodeRadius)
            .attr('data-depth', d => d.depth);

        // Add labels
        node.append('text')
            .text(d => d.name)
            .attr('dy', d => d.isTarget ? 40 : 30)
            .style('display', this.config.showLabels ? 'block' : 'none')
            .style('font-size', d => d.isTarget ? '14px' : '12px')
            .style('font-weight', d => d.isTarget ? 'bold' : 'normal');

        // Add tooltips
        node.append('title')
            .text(d => `${d.name}\n${d.kind} in ${d.package || 'external'}`);

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
        // Create nodes array (target + all other nodes)
        this.nodes = [this.data.target, ...this.data.nodes];

        // Filter external nodes if needed
        if (!this.config.showExternal) {
            this.nodes = this.nodes.filter(n => !n.external);
        }

        // Create links from edges
        this.links = this.data.edges.map(edge => ({
            source: edge.from,
            target: edge.to,
            type: edge.type,
            depth: edge.depth,
            label: edge.label
        }));

        // Filter links to only include visible nodes
        const nodeIds = new Set(this.nodes.map(n => n.name));
        this.links = this.links.filter(l =>
            nodeIds.has(l.source) && nodeIds.has(l.target)
        );
    }

    showNodeDetails(node) {
        const codeContent = document.getElementById('code-content');
        const codeTitle = document.getElementById('code-title');

        // Build details HTML
        let html = '<div class="node-details">';

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
            html += `<div class="detail-row">
                <span class="detail-label">File:</span>
                <span class="detail-value">${fileName}:${node.line}</span>
            </div>`;
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

        // Add documentation if available
        if (node.doc && this.config.showDocs) {
            html += `<div class="node-doc">${this.escapeHtml(node.doc)}</div>`;
        }

        // Add code if available
        if (node.code) {
            html += '<h3>Code</h3>';
            html += `<div class="code-block"><pre>${this.escapeHtml(node.code)}</pre></div>`;
        } else if (node.external) {
            html += '<div class="empty-state">External symbol - code not available</div>';
        }

        codeTitle.textContent = `${node.name} (${node.kind})`;
        codeContent.innerHTML = html;
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
    new GoScopeVisualizer();
});
