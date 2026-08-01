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
    console.error("Usage: node oracle_cluster2.mjs <input.json>");
    process.exit(1);
}

const inputData = fs.readFileSync(inputFile, 'utf-8');
if (!inputData || !inputData.trim()) {
    process.exit(0);
}

const cases = JSON.parse(inputData);
const results = cases.map(c => {
    try {
        let xVal, yVal;
        
        // Parse Operand X
        if (c.xType === 'string') {
            xVal = JSBI.BigInt(c.x);
        } else if (c.xType === 'number') {
            let n = c.x;
            if (c.x === 'Infinity' || c.x === '+Infinity') n = Infinity;
            else if (c.x === '-Infinity') n = -Infinity;
            else if (c.x === 'NaN') n = NaN;
            xVal = JSBI.BigInt(n);
        } else {
            xVal = JSBI.BigInt(c.x);
        }

        // Parse Operand Y
        if (c.yType === 'number') {
            let n = c.y;
            if (c.y === 'Infinity' || c.y === '+Infinity') n = Infinity;
            else if (c.y === '-Infinity') n = -Infinity;
            else if (c.y === 'NaN') n = NaN;
            
            // Compare BigInt X against Float Y directly via JSBI.__compareToDouble
            const comp = JSBI.__compareToDouble(xVal, n);
            const eq = JSBI.__equalToNumber(xVal, n);
            const ne = !eq;
            const lt = comp < 0;
            const le = comp <= 0;
            const gt = comp > 0;
            const ge = comp >= 0;
            return { status: 'OK', comp, eq, ne, lt, le, gt, ge };
        } else {
            if (c.yType === 'string') {
                yVal = JSBI.BigInt(c.y);
            } else {
                yVal = JSBI.BigInt(c.y);
            }

            const comp = JSBI.__compareToBigInt(xVal, yVal);
            const eq = JSBI.equal(xVal, yVal);
            const ne = JSBI.notEqual(xVal, yVal);
            const lt = JSBI.lessThan(xVal, yVal);
            const le = JSBI.lessThanOrEqual(xVal, yVal);
            const gt = JSBI.greaterThan(xVal, yVal);
            const ge = JSBI.greaterThanOrEqual(xVal, yVal);
            return { status: 'OK', comp, eq, ne, lt, le, gt, ge };
        }
    } catch (e) {
        return { status: 'ERR', errName: e.name, errMessage: e.message };
    }
});

console.log(JSON.stringify(results));
