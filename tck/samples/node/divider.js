module.exports = x => {
    if (x === 0) { // would otherwise return NaN without crashing
        throw new Error("Division by zero!");
    }
    return 100 / x;
};