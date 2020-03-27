const {Readable, Transform} = require("stream");

// do not use in production - no backpressure limitation in place...
class Zipper extends Readable {

    constructor(...streams) {
        super({objectMode: true});
        this.waitingForData = false;
        this.arrays = streams.map(() => []);
        streams.forEach((stream, index) => {
            stream.on('data', (chunk) => {
                const array = this.arrays[index];
                array.push(chunk);
                if (this.waitingForData) {
                    this._read(array.length);
                }
            });
        })
    }

    _read(size) {
        const bound = Math.min(size, ...this.arrays.map(a => a.length));
        if (bound === 0) {
            this.waitingForData = true;
            return;
        }
        this.waitingForData = false;
        this.arrays
            .map(a => a.splice(0, bound))
            .reduce((prev, curr) => {
                const result = [];
                for (let i = 0; i < bound; i++) {
                    const previous = (Array.isArray(prev[i])) ? prev[i] : [prev[i]];
                    result.push([...previous, curr[i]]);
                }
                return result
            })
            .forEach((result) => {
                this.push(result);
            });
    }
}

class Repeater extends Transform {
    constructor() {
        super({objectMode: true});
    }

    _transform(tuple, _, callback) {
        const repetition = tuple[0];
        for (let i = 0; i < repetition; i++) {
            this.push(tuple[1]);
        }
        callback();
    }
}

module.exports = (inputs, outputs) => {
    new Zipper(inputs.$order[0], inputs.$order[1])
        .pipe(new Repeater())
        .pipe(outputs.$order[0]);
};
module.exports.$interactionModel = 'node-streams';
