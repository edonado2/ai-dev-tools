// scripts/download-binary.js
const { platform, arch } = process;
const fs = require('fs');
const path = require('path');
const https = require('https');
const { execSync } = require('child_process');

const version = require('../package.json').version;
const base = `https://github.com/ernedo/ai-dev-tools/releases/download/v${version}`;
const map = { darwin: 'Darwin', linux: 'Linux', win32: 'Windows' };
const archMap = { x64: 'x86_64', arm64: 'arm64' };
const ext = platform === 'win32' ? '.zip' : '.tar.gz';
const url = `${base}/ai-dev-tools_${version}_${map[platform]}_${archMap[arch]}${ext}`;
const destDir = path.join(__dirname, '../lib');
const destPath = path.join(destDir, platform === 'win32' ? 'ai-dev.exe' : 'ai-dev');

async function downloadBinary() {
  if (!fs.existsSync(destDir)) {
    fs.mkdirSync(destDir, { recursive: true });
  }

  console.log(`Downloading ${url}...`);

  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(path.join(destDir, `archive${ext}`));
    https.get(url, (response) => {
      if (response.statusCode === 302 || response.statusCode === 301) {
        // Follow redirect
        https.get(response.headers.location, (res) => {
          res.pipe(file);
          file.on('finish', () => {
            file.close();
            extractArchive();
          });
        }).on('error', reject);
      } else if (response.statusCode === 200) {
        response.pipe(file);
        file.on('finish', () => {
          file.close();
          extractArchive();
        });
      } else {
        reject(new Error(`Failed to download: ${response.statusCode}`));
      }
    }).on('error', reject);

    function extractArchive() {
      const archivePath = path.join(destDir, `archive${ext}`);
      try {
        if (platform === 'win32') {
          execSync(`powershell -Command "Expand-Archive -Force -Path '${archivePath}' -DestinationPath '${destDir}'"`);
        } else {
          execSync(`tar -xzf "${archivePath}" -C "${destDir}"`);
        }
        fs.chmodSync(destPath, 0o755);
        fs.unlinkSync(archivePath);
        console.log(`Binary installed to ${destPath}`);
        resolve(destPath);
      } catch (err) {
        reject(err);
      }
    }
  });
}

if (require.main === module) {
  downloadBinary().catch((err) => {
    console.error('Download failed:', err.message);
    process.exit(1);
  });
}

module.exports = { downloadBinary };