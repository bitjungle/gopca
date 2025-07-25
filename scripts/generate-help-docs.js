#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

// Load help content
const helpContentPath = path.join(__dirname, '../cmd/gopca-desktop/frontend/src/help/help-content.json');
const helpContent = JSON.parse(fs.readFileSync(helpContentPath, 'utf8'));

// Generate Markdown documentation
function generateMarkdown() {
    let markdown = '# GoPCA Help Documentation\n\n';
    markdown += 'This documentation is automatically generated from the help system.\n\n';
    
    // Group help items by category
    const byCategory = {};
    Object.entries(helpContent.help).forEach(([key, item]) => {
        const category = item.category;
        if (!byCategory[category]) {
            byCategory[category] = [];
        }
        byCategory[category].push({ key, ...item });
    });
    
    // Generate sections by category
    Object.entries(helpContent.categories).forEach(([catKey, catInfo]) => {
        markdown += `## ${catInfo.name}\n\n`;
        markdown += `${catInfo.description}\n\n`;
        
        const items = byCategory[catKey] || [];
        items.forEach(item => {
            markdown += `### ${item.title}\n\n`;
            markdown += `${item.text}\n\n`;
        });
    });
    
    return markdown;
}

// Generate CLI help text
function generateCLIHelp() {
    let cliHelp = 'GoPCA CLI Help\n\n';
    
    // Filter items relevant to CLI
    const cliRelevant = ['num-components', 'pca-method', 'mean-center', 'standard-scale', 
                        'robust-scale', 'missing-strategy', 'snv', 'l2-norm'];
    
    Object.entries(helpContent.help).forEach(([key, item]) => {
        if (cliRelevant.includes(key)) {
            cliHelp += `${item.title}:\n  ${item.text}\n\n`;
        }
    });
    
    return cliHelp;
}

// Generate JSON for other tools
function generateJSON() {
    const output = {
        version: helpContent.metadata.version,
        generated: new Date().toISOString(),
        help: helpContent.help,
        categories: helpContent.categories
    };
    
    return JSON.stringify(output, null, 2);
}

// Main execution
const outputDir = path.join(__dirname, '../docs/generated');
if (!fs.existsSync(outputDir)) {
    fs.mkdirSync(outputDir, { recursive: true });
}

// Generate all formats
fs.writeFileSync(path.join(outputDir, 'help-guide.md'), generateMarkdown());
fs.writeFileSync(path.join(outputDir, 'cli-help.txt'), generateCLIHelp());
fs.writeFileSync(path.join(outputDir, 'help-content.json'), generateJSON());

console.log('Help documentation generated successfully!');
console.log(`  - Markdown: docs/generated/help-guide.md`);
console.log(`  - CLI Help: docs/generated/cli-help.txt`);
console.log(`  - JSON: docs/generated/help-content.json`);