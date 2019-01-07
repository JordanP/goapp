#!/bin/sh -e

: ${COVER_RESULTS:=/tmp/cover-results}

mkdir -p $COVER_RESULTS
touch $COVER_RESULTS/coverage.tmp

echo 'mode: atomic' > $COVER_RESULTS/coverage.cover
for d in $(go list ./... | grep -v vendor); do
    go test -coverprofile=$COVER_RESULTS/coverage.tmp --coverpkg=$(go list ./... | paste -sd "," -) $d 2>&1 | grep -v "no packages being tested depend"
    if [ -f $COVER_RESULTS/coverage.tmp ]; then
        tail -n +2 $COVER_RESULTS/coverage.tmp >> $COVER_RESULTS/coverage.cover
        rm ${COVER_RESULTS}/coverage.tmp
    fi
done
go tool cover -html=${COVER_RESULTS}/coverage.cover -o ${COVER_RESULTS}/coverage.html

echo
echo "To open the html coverage file use one of the following commands:"
echo "open $COVER_RESULTS/coverage.html"
echo "xdg-open .tmp/cover-results/coverage.html"
