# metaphone3 - a sound-a-like index for names
TODO: description

## Usage
TODO: code sample

## Basis for algorithm
The reference implementation of metaphone3 in Java can be found [here](https://github.com/OpenRefine/OpenRefine/blob/master/main/src/com/google/refine/clustering/binning/Metaphone3.java).

## Differences from v2.1.3 Java Implementation
- Fix ROBILL
- Fix lengths for very long words where certain situations would cause primary 
    to get too long and the secondary would get truncated.  (e.g. Villafranca when EncodeVowel is true)
