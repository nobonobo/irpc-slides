import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const rootDir = path.resolve(__dirname, '..');
const distDir = path.resolve(rootDir, 'dist');

console.log('🚀 Building standalone dist/index.html...');

// 1. Process index.html -> dist/index.html
const indexPath = path.resolve(rootDir, 'index.html');
if (fs.existsSync(indexPath)) {
	let html = fs.readFileSync(indexPath, 'utf-8');

	// Replace dist/ references with relative paths (./)
	// Example: href="dist/reset.css" -> href="./reset.css"
	//          src="dist/reveal.js" -> src="./reveal.js"
	//          data-markdown="dist/contents.md" -> data-markdown="./contents.md"
	html = html.replace(/(href|src|data-markdown)=["']dist\//g, '$1="./');

	const distIndexPath = path.resolve(distDir, 'index.html');
	fs.writeFileSync(distIndexPath, html, 'utf-8');
	console.log(`✅ Generated: ${distIndexPath}`);
} else {
	console.error('❌ index.html not found in root directory.');
}

// 2. Process dist/contents.md if present (replace internal dist/ references if any)
const distContentsPath = path.resolve(distDir, 'contents.md');
if (fs.existsSync(distContentsPath)) {
	let contents = fs.readFileSync(distContentsPath, 'utf-8');
	if (contents.includes('dist/')) {
		contents = contents.replace(/\bdist\//g, './');
		fs.writeFileSync(distContentsPath, contents, 'utf-8');
		console.log(`✅ Updated paths in: ${distContentsPath}`);
	}
}

console.log('✨ Build script completed successfully.');
