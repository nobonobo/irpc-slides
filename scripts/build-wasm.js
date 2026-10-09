import { execSync } from 'child_process';
import { dirname, resolve } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
const root = resolve(__dirname, '..');

try {
  console.log('Building WASM (GOOS=js GOARCH=wasm)...');
  execSync('go build -o public/main.wasm .', {
    cwd: root,
    stdio: 'inherit',
    env: {
      ...process.env,
      GOOS: 'js',
      GOARCH: 'wasm'
    }
  });
  console.log('WASM build completed: public/main.wasm');
} catch (err) {
  console.error('Failed to build WASM:', err);
  process.exit(1);
}
