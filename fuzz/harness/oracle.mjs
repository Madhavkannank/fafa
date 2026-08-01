import fs from 'fs';
import path from 'path';
import { fileURLToPath, pathToFileURL } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const jsbiPath = path.resolve(__dirname, '../../jsbi/dist/jsbi.mjs');
const JSBI = (await import(pathToFileURL(jsbiPath).href)).default;

const inputFile = process.argv[2];
if (!inputFile) {
    console.error("Usage: node oracle.mjs <input.json>");
    process.exit(1);
}

const inputData = fs.readFileSync(inputFile, 'utf-8');
if (!inputData || !inputData.trim()) {
    process.exit(0);
}

const cases = JSON.parse(inputData);
const results = cases.map(c => {
    try {
        let b;
        if (c.type === 'string') {
            if (c.radix && c.radix !== 0) {
                b = JSBI.__fromString(c.value, c.radix);
                if (b === null) {
                    throw new SyntaxError('Cannot convert ' + c.value + ' to a BigInt');
                }
            } else {
                b = JSBI.BigInt(c.value);
            }
        } else if (c.type === 'number') {
            let numVal = c.value;
            if (c.value === 'Infinity' || c.value === '+Infinity') numVal = Infinity;
            else if (c.value === '-Infinity') numVal = -Infinity;
            else if (c.value === 'NaN') numVal = NaN;
            b = JSBI.BigInt(numVal);
        } else if (c.type === 'boolean') {
            b = JSBI.BigInt(c.value);
        } else {
            b = JSBI.BigInt(c.value);
        }
        return { status: 'OK', sign: b.sign, len: b.length, digits: Array.from(b) };
    } catch (e) {
        return { status: 'ERR', errName: e.name, errMessage: e.message };
    }
});

console.log(JSON.stringify(results));
