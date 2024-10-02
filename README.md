# Licenses

Simple program that extracts some license data from a CycloneDX type SBOM,
in JSON format.

This is then output to a file based on a template. Default template included in
the box.

**Note:** this helper program does *not* do any scanning, it just ingests an SBOM.

## Why?

So why create this? Simple, I was required to provide a simple, human-readable
file about third-party licenses for another project. Most tools I found were too
cumbersome, buggy, complex, etc. Since I already had a CycloneDX style SBOM, why
not re-use...

## Licensing

This software is made available under the [MPL-2.0](https://choosealicense.com/licenses/mpl-2.0/) license. The full details are available from the [LICENSE](/LICENSE) file.

Copyright (C) 2024  Martijn van der Kleijn