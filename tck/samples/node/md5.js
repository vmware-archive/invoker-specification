const crypto = require("crypto")

module.exports = x => crypto.createHash("md5").update(x).digest("hex")