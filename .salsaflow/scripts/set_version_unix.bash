#!/bin/bash

set -e

version="$1"

# Make sure the version string is in the first argument.
if [ -z "$version" ]; then
	echo "No version string supplied!" 1>&2
	exit 1
fi

# Create the metadata package directory.
mkdir -p app/metadata

# Write the version string into version.go so that it gets
# compiled into the resulting executable.
cat > app/metadata/version.go <<-EOF
package metadata

/*****************************************
 * This file was generated by SalsaFlow. *
 * Please do not modify it manually.     *
 *****************************************/

const Version = "$version"
EOF

# Stage the version file to be sure it is committed.
git add app/metadata/version.go
