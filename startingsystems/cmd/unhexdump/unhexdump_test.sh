#!/usr/bin/env bash
../../bin/echo "to the last I grapple with thee; from hell's heart I stab at thee; for hate's sake I spit my last breath at thee" > moby.txt
../../bin/shexdump moby.txt > moby.hex
../../bin/unhexdump moby.hex > moby2.txt
diff -s moby.txt moby2.txt