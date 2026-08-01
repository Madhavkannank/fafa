import fs from 'fs';
import path from 'path';
import { fileURLToPath, pathToFileURL } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// IMPORT LINE CONFIRMATION: Importing real compiled ESM distribution from reference GoogleChromeLabs/jsbi
const jsbiPath = path.resolve(__dirname, '../../jsbi/dist/jsbi.mjs');
const JSBI = (await import(pathToFileURL(jsbiPath).href)).default;

const inputFile = process.argv[2];
if (!inputFile) {
    console.error("Usage: node oracle_cluster4.mjs <input.json>");
    process.exit(1);
}

const inputData = fs.readFileSync(inputFile, 'utf-8');
if (!inputData || !inputData.trim()) {
    process.exit(0);
}

const cases = JSON.parse(inputData);
const results = cases.map(c => {
    try {
        const xVal = JSBI.BigInt(c.x);
        const yVal = JSBI.BigInt(c.y);

        const mulRes = JSBI.multiply(xVal, yVal);

        return {
            status: 'OK',
            mulSign: mulRes.sign,
            mulLen: mulRes.length,
            mulDigits: Array.from(mulRes)
        };
    } catch (e) {
        return { status: 'ERR', errName: e.name, errMessage: e.message };
    }
});

console.log(JSON.stringify(results));
