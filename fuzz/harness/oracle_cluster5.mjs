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
    console.error("Usage: node oracle_cluster5.mjs <input.json>");
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

        let divRes = null;
        let remRes = null;
        let divErr = null;
        let remErr = null;

        try {
            divRes = JSBI.divide(xVal, yVal);
        } catch (e) {
            divErr = { errName: e.name, errMessage: e.message };
        }

        try {
            remRes = JSBI.remainder(xVal, yVal);
        } catch (e) {
            remErr = { errName: e.name, errMessage: e.message };
        }

        if (divErr || remErr) {
            return {
                status: 'ERR',
                errName: (divErr || remErr).errName,
                errMessage: (divErr || remErr).errMessage
            };
        }

        return {
            status: 'OK',
            divSign: divRes.sign,
            divLen: divRes.length,
            divDigits: Array.from(divRes),
            remSign: remRes.sign,
            remLen: remRes.length,
            remDigits: Array.from(remRes)
        };
    } catch (e) {
        return { status: 'ERR', errName: e.name, errMessage: e.message };
    }
});

console.log(JSON.stringify(results));
