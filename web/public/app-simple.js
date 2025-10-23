// Simplified version - Add code links AFTER Prism, not before
// This is a monkey-patch to test if it works

// Override the showNodeDetails method after page loads
setTimeout(() => {
    const viz = window.goScopeViz;
    if (!viz) return;

    const originalShow = viz.showNodeDetails.bind(viz);

    viz.showNodeDetails = function(node) {
        // Call original
        originalShow(node);

        // Now add links after Prism has done its thing
        const codeBlock = document.querySelector('.code-block code');
        if (!codeBlock || !this.nodes) return;

        console.log('Adding links after Prism...');

        // Build symbol map from all symbols (not just file nodes)
        const symbolMap = new Map();
        const allSymbols = this.symbols || this.nodes || [];
        allSymbols.forEach(n => {
            if (n.name && !n.external && n.file) {
                symbolMap.set(n.name, n);
            }
        });

        console.log('Found', symbolMap.size, 'symbols to link');

        // Sort by length (longest first) to match longer names first
        const symbols = Array.from(symbolMap.keys()).sort((a, b) => b.length - a.length);

        let linkCount = 0;

        // Work with DOM nodes to avoid breaking Prism's spans
        const walker = document.createTreeWalker(
            codeBlock,
            NodeFilter.SHOW_TEXT,
            null
        );

        const textNodes = [];
        let currentNode;
        while (currentNode = walker.nextNode()) {
            textNodes.push(currentNode);
        }

        // Process each text node
        textNodes.forEach(textNode => {
            let text = textNode.textContent;
            const matches = [];

            // Find all matches in this text node
            symbols.forEach(symbolName => {
                const escaped = symbolName.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
                const pattern = new RegExp(`\\b${escaped}\\b`, 'g');
                let match;

                while ((match = pattern.exec(text)) !== null) {
                    matches.push({
                        start: match.index,
                        end: match.index + match[0].length,
                        text: match[0],
                        node: symbolMap.get(symbolName)
                    });
                }
            });

            if (matches.length === 0) return;

            // Sort matches by position and remove overlaps (keep first match)
            matches.sort((a, b) => a.start - b.start);
            const nonOverlapping = [];
            let lastEnd = 0;

            matches.forEach(m => {
                if (m.start >= lastEnd) {
                    nonOverlapping.push(m);
                    lastEnd = m.end;
                }
            });

            if (nonOverlapping.length === 0) return;

            // Build fragments
            const fragments = [];
            let lastIndex = 0;

            nonOverlapping.forEach(m => {
                // Add text before match
                if (m.start > lastIndex) {
                    fragments.push(document.createTextNode(text.substring(lastIndex, m.start)));
                }

                // Add link span
                const link = document.createElement('span');
                link.className = 'code-link';
                link.setAttribute('data-node-id', m.node.id);
                link.setAttribute('data-node-name', m.node.name);
                link.setAttribute('title', `Click to view ${m.node.name}`);
                link.textContent = m.text;
                fragments.push(link);
                linkCount++;

                lastIndex = m.end;
            });

            // Add remaining text
            if (lastIndex < text.length) {
                fragments.push(document.createTextNode(text.substring(lastIndex)));
            }

            // Replace the text node with our fragments
            const parent = textNode.parentNode;
            fragments.forEach(frag => {
                parent.insertBefore(frag, textNode);
            });
            parent.removeChild(textNode);
        });

        console.log('Created', linkCount, 'links');

        // Attach click handlers
        const links = codeBlock.querySelectorAll('.code-link');
        console.log('Attached', links.length, 'click handlers');

        links.forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                e.stopPropagation();
                const nodeName = link.getAttribute('data-node-name');
                console.log('Clicked:', nodeName);

                // Find the symbol first
                const allSymbols = this.symbols || this.nodes || [];
                const targetSymbol = allSymbols.find(n => n.name === nodeName);

                if (targetSymbol && targetSymbol.file) {
                    // Find the file node that contains this symbol
                    const fileNode = this.nodes.find(n => n.kind === 'file' && n.id === targetSymbol.file);
                    if (fileNode) {
                        this.showNodeDetails(fileNode);
                        this.highlightNode(fileNode);
                    } else {
                        console.warn('File node not found for:', targetSymbol.file);
                    }
                } else {
                    console.warn('Symbol not found:', nodeName);
                }
            });
        });
    };
}, 1000);
