//Package metaphone3 is a Go implementation of the Metaphone 3 algorithm.
// Metaphone 3 is designed to return an *approximate* phonetic key (and an alternate
// approximate phonetic key when appropriate) that should be the same for English
// words, and most names familiar in the United States, that are pronounced *similarly*.
// The key value is *not* intended to be an *exact* phonetic, or even phonemic,
// representation of the word. This is because a certain degree of 'fuzziness' has
// proven to be useful in compensating for variations in pronunciation, as well as
// misheard pronunciations. For example, although americans are not usually aware of it,
// the letter 's' is normally pronounced 'z' at the end of words such as "sounds".
//
// The 'approximate' aspect of the encoding is implemented according to the following rules:
//
// (1) All vowels are encoded to the same value - 'A'. If the parameter encodeVowels
// is set to false, only *initial* vowels will be encoded at all. If encodeVowels is set
// to true, 'A' will be encoded at all places in the word that any vowels are normally
// pronounced. 'W' as well as 'Y' are treated as vowels. Although there are differences in
// the pronunciation of 'W' and 'Y' in different circumstances that lead to their being
// classified as vowels under some circumstances and as consonants in others, for the purposes
// of the 'fuzziness' component of the Soundex and Metaphone family of algorithms they will
// be always be treated here as vowels.
//
// (2) Voiced and un-voiced consonant pairs are mapped to the same encoded value. This
// means that:
// 'D' and 'T' -> 'T'
// 'B' and 'P' -> 'P'
// 'G' and 'K' -> 'K'
// 'Z' and 'S' -> 'S'
// 'V' and 'F' -> 'F'
//
// - In addition to the above voiced/unvoiced rules, 'CH' and 'SH' -> 'X', where 'X'
// represents the "-SH-" and "-CH-" sounds in Metaphone 3 encoding.
//
// - Also, the sound that is spelled as "TH" in English is encoded to '0' (zero symbol). (Although
// Americans are not usually aware of it, "TH" is pronounced in a voiced (e.g. "that") as
// well as an unvoiced (e.g. "theater") form, which are naturally mapped to the same encoding.)
//
// The encodings in this version of Metaphone 3 are according to pronunciations common in the
// United States. This means that they will be inaccurate for consonant pronunciations that
// are different in the United Kingdom, for example "tube" -> "CHOOBE" -> XAP rather than american TAP.
//
// Metaphone 3 was preceded by Soundex, patented in 1919, and Metaphone and Double Metaphone,
// developed by Lawrence Philips. All of these algorithms resulted in a significant number of
// incorrect encodings. Metaphone3 was tested against a database of about 100 thousand English words,
// names common in the United States, and non-English words found in publications in the United States,
// with an emphasis on words that are commonly mispronounced, prepared by the Moby Words website,
// but with the Moby Words 'phonetic' encodings algorithmically mapped to Double Metaphone encodings.
// Metaphone3 increases the accuracy of encoding of english words, common names, and non-English
// words found in american publications from the 89% for Double Metaphone, to over 98%.
//
package metaphone3
