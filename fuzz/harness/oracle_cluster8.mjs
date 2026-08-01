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
    const op = input.op; // 'asIntN', 'asUintN'
    const bits = input.bits;
    const x = JSBI.BigInt(input.x);
    let result;

    if (op === 'asIntN') {
      result = JSBI.asIntN(bits, x);
    } else if (op === 'asUintN') {
      result = JSBI.asUintN(bits, x);
    } else {
      throw new Error('Unknown op: ' + op);
    }

    const digits = [];
    for (let i = 0; i < result.length; i++) {
      digits.push(result.__digit(i));
    }

    console.log(JSON.stringify({
      sign: result.sign,
      length: result.length,
      digits: digits,
      str: result.toString(10),
      err: ""
    }));
  } catch (e) {
    console.log(JSON.stringify({
      sign: false,
      length: 0,
      digits: [],
      str: "",
      err: e.message || String(e)
    }));
  }
});
