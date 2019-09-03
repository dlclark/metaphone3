# metaphone3 - a sound-a-like index for names
Metaphone3 is a more accurate version of the original Soundex algorithm.  It's designed so that similar-sounding words in American English share the same keys.  For example Smith, Smyth, Smithe, Smythe all encode to `SM0` primary and `XMT` alt.  Whereas Schmidt encodes to `XMT` primary with no secondary.  

Searching for matches where either primary or secondary matches will give the best results.

You can read more about Metaphone on [Wikipedia](https://en.wikipedia.org/wiki/Metaphone).

## Usage
Basic usage of the encoder looks like this:
```go
	e := &metaphone3.Encoder{}
	prim, second := e.Encode("Smith")
```

An `Encoder` is designed to be re-used to reduce memory pressure at scale and has three settable options.  An `Encoder` is not thread-safe so it is not safe to use one `Encoder` across goroutines.  If you're comparing values you *must* use the exact same options.


| Option | Type | Default | Purpose |
| --- | --- | --- | --- |
| `EncodeExact` | `bool` | `false` | Setting `EncodeExact` to `true` will tighten the output so that certain sounds will be differentiated.  E.g. more separation between hard "G" sounds and hard "K" sounds. |
| `EncodeVowels` | `bool` | `false` | Setting `EncodeVowels` to `true` will include non-first-letter vowel sounds in the output.  By default only consonent sounds are included. |
| `MaxLength` | `int` | `metaphone3.DefaultMaxLength` | This limits the output of long words and is useful to reduce the cycles and memory spent on processing long words. |
| `metaphone3.DefaultMaxLength` | `int` | 8 | If `MaxLength` is `0` (or negative) then it defaults as `metaphone3.DefaultMaxLength`, which starts as `8` (like the java implementation). |

Additional usage details available in the [godocs](https://godoc.org/github.com/dlclark/metaphone3).

## Basis for algorithm
The reference implementation of metaphone3 in Java can be found [here](https://github.com/OpenRefine/OpenRefine/blob/master/main/src/com/google/refine/clustering/binning/Metaphone3.java).

## Differences from v2.1.3 Java Implementation
- Fix ROBILL
- Fix lengths for very long words where certain situations would cause the primary or secondary
    to get too long and the other would get truncated.  (e.g. Villafranca when EncodeVowel is true)
- Fix JAKOB
- Fix ending CIAS and CIOS (e.g. MECIAS)
- Fix words starting with HARGER
- Fix SUPERNODE (prevent D from being silent)