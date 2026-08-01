import fs from 'fs';
import path from 'path';
import { fileURLToPath, pathToFileURL } from 'url';
import readline from 'readline';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Importing compiled ESM distribution from GoogleChromeLabs/jsbi
const jsbiPath = path.resolve(__dirname, '../../jsbi/dist/jsbi.mjs');
const JSBI = (await import(pathToFileURL(jsbiPath).href)).default;

/**
 * Oracle for Cluster 6 (Shifts)
 * Protocol: Reads JSON lines from stdin, outputs JSON lines to stdout.
 * Input format: { op: "leftShift" | "signedRightShift", x: string, y: string }
 * Output format: { result: string, err?: string, digits: number[], sign: boolean, length: number }
 */
const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout,
  terminal: false
});

rl.on('line', (line) => {
  if (!line.trim()) return;
  try {
    const req = JSON.parse(line);
    const x = JSBI.BigInt(req.x);
    const y = JSBI.BigInt(req.y);
    let res;
    if (req.op === 'leftShift') {
      res = JSBI.leftShift(x, y);
    } else if (req.op === 'signedRightShift') {
      res = JSBI.signedRightShift(x, y);
    } else {
      throw new Error(`Unknown op: ${req.op}`);
    }

    const digits = [];
    for (let i = 0; i < res.length; i++) {
      digits.push(res.__digit(i));
    }

    const resp = {
      result: res.toString(),
      sign: res.sign,
      length: res.length,
      digits: digits
    };
    console.log(JSON.stringify(resp));
  } catch (err) {
    console.log(JSON.stringify({ err: err.message }));
  }
});
