import { createCanvas } from 'canvas';
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

function generateIcon(size, filename) {
  const canvas = createCanvas(size, size);
  const ctx = canvas.getContext('2d');

  // Background
  ctx.fillStyle = '#1976d2';
  ctx.fillRect(0, 0, size, size);

  // Text
  ctx.fillStyle = '#ffffff';
  ctx.font = `bold ${size * 0.3}px Arial`;
  ctx.textAlign = 'center';
  ctx.textBaseline = 'middle';
  ctx.fillText('MTX', size / 2, size / 2);

  // Save to file
  const buffer = canvas.toBuffer('image/png');
  fs.writeFileSync(path.join(__dirname, 'public', filename), buffer);
  console.log(`Generated ${filename}`);
}

// Generate both icons
generateIcon(192, 'icon-192.png');
generateIcon(512, 'icon-512.png'); 