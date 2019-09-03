# Function Invoker Specification and Build Guide
This specification defines the role of a *function invoker* and how it interacts with other components in the riff ecosystem, namely a *streaming processor* and a *function*.
The aim of this document is to formaly describe the protocol an invoker must adhere to, as well as prescribe a recommended way to build and package such an invoker.

## Conventions
The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and "OPTIONAL" are to be interpreted as described in [RFC 2119](http://tools.ietf.org/html/rfc2119).

## Sections
* [Role of an invoker](introduction.md)
* [Streaming Specification](streaming.md)
* [Request/Reply Specification](request-reply.md)
* [Packaging an invoker for riff](packaging.md)
   * use of buildpacks
   * contract with riff cli: riff.toml
   * how to get "http for free"
* [Glossary](glossary.md)