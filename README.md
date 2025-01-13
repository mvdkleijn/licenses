# Licenses

Simple program that extracts some license data from a CycloneDX type SBOM,
in JSON format.

This is then output to a file based on a template. Default template included in
the box.

Optionally, using the --validate command line option, this program can make a
simplistic judgement on whether the licenses of dependencies are compatible with
the main application license. This is done using the compatibility.yaml source file
which also lists the relevant disclaimer and reasons for (in)compatibility.

**Note:** this helper program does *not* do any scanning, it just ingests an SBOM.

## Why?

So why create this? Simple, I was required to provide a simple, human-readable
file about third-party licenses for another project. Most tools I found were too
cumbersome, buggy, complex, etc. Since I already had a CycloneDX style SBOM, why
not re-use...

## Disclaimer

Licensing compatibility can be nuanced, especially for combined or derivative works.
Always refer to the official license texts, their appendices (if any), and make use of
additional legal guidance to ensure you’re applying compatibility rules correctly.

This piece of software does its best to be accurrate but gives no guarantees nor
warranties of any kind.

## Licensing

This software is made available under the [MPL-2.0](https://choosealicense.com/licenses/mpl-2.0/) license.
The full details are available from the [LICENSE](/LICENSE) file.

Copyright (C) 2024-2025  Martijn van der Kleijn