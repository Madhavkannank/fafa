import readline from 'readline';
import JSBI from '../../jsbi/dist/jsbi.mjs';

const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout,
  terminal: false
});

rl.on('line', (line) => {
  if (!line.trim()) return;
  try {
    const input = JSON.parse(line);
    const radix = input.radix;
    const x = JSBI.BigInt(input.x);
    const result = x.toString(radix);
    console.log(JSON.stringify({ str: result, err: "" }));
  } catch (e) {
    console.log(JSON.stringify({ str: "", err: e.message || String(e) }));
  }
});
