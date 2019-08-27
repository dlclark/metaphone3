package metaphone3

import (
	"fmt"
	"unicode"
)

// debug flag to output additional data during encoding
var debug = false

// DefaultMaxLength is the max number of runes in a result when not specified in the encoder
var DefaultMaxLength = 8

// Encoder is a metaphone3 encoder that contains options and state for encoding.  It is not
// safe to use across goroutines.
type Encoder struct {
	// EncodeVowels determines if Metaphone3 will encode non-initial vowels. However, even
	// if there are more than one vowel sound in a vowel sequence (i.e.
	// vowel diphthong, etc.), only one 'A' will be encoded before the next consonant or the
	// end of the word.
	EncodeVowels bool

	// EncodeExact controls if Metaphone3 will encode consonants as exactly as possible.
	// This does not include 'S' vs. 'Z', since americans will pronounce 'S' at the
	// at the end of many words as 'Z', nor does it include "CH" vs. "SH". It does cause
	// a distinction to be made between 'B' and 'P', 'D' and 'T', 'G' and 'K', and 'V'
	// and 'F'.
	EncodeExact bool

	// The max allowed length of the output metaphs, if <= 0 then the DefaultMaxLength is used
	MaxLength int

	in                 []rune
	idx                int
	lastIdx            int
	primBuf, secondBuf []rune
	flagAlInversion    bool
}

// Encode takes in a string and returns primary and secondary metaphones.
// Both will be blank if given a blank input, and secondary can be blank
// if there's only one metaphone.
func (e *Encoder) Encode(in string) (primary, secondary string) {
	if in == "" {
		return "", ""
	}

	if e.MaxLength <= 0 {
		e.MaxLength = DefaultMaxLength
	}

	e.flagAlInversion = false

	// setup our input buffer and to-upper everything
	e.in = make([]rune, 0, len(in))
	for _, r := range in {
		e.in = append(e.in, unicode.ToUpper(r))
	}
	e.lastIdx = len(e.in) - 1

	e.primBuf = primeBuf(e.primBuf, e.MaxLength)
	e.secondBuf = primeBuf(e.secondBuf, e.MaxLength)

	// lets go rune-by-rune through the input string
	for e.idx = 0; e.idx < len(e.in); e.idx++ {

		// double check our output buffers, if they're full then we're done
		// we're not checking exact "=" just be compat with the reference java implementation
		// that means our buffers could be longer than MaxLength by a bit
		if len(e.primBuf) >= e.MaxLength && len(e.secondBuf) >= e.MaxLength {
			break
		}

		if debug {
			fmt.Printf("Processing %v\n", string(e.in[e.idx]))
		}

		switch c := e.in[e.idx]; c {
		case 'B':
			e.encodeB()
		case 'ß', 'Ç':
			e.metaphAdd('S')
		case 'C':
			e.encodeC()
		case 'D':
			e.encodeD()
		case 'F':
			e.encodeF()
		case 'G':
			e.encodeG()
		case 'H':
			e.encodeH()
		case 'J':
			e.encodeJ()
		case 'K':
			e.encodeK()
		case 'L':
			e.encodeL()
		case 'M':
			e.encodeM()
		case 'N':
			e.encodeN()
		case 'Ñ':
			e.metaphAdd('N')
		case 'P':
			e.encodeP()
		case 'Q':
			e.encodeQ()
		case 'R':
			e.encodeR()
		case 'S':
			e.encodeS()
		case 'T':
			e.encodeT()
		case 'Ð', 'Þ':
			e.metaphAdd('0')
		case 'V':
			e.encodeV()
		case 'W':
			e.encodeW()
		case 'X':
			e.encodeX()
		case '\uC28A':
			//wat?
			e.metaphAdd('X')
		case '\uC28E':
			//wat?
			e.metaphAdd('S')
		case 'Z':
			e.encodeZ()
		default:
			if isVowel(c) {
				e.encodeVowels()
			}
		}
	}

	// trim our buffers if needed
	if len(e.primBuf) > e.MaxLength {
		e.primBuf = e.primBuf[:e.MaxLength]
	}
	if len(e.secondBuf) > e.MaxLength {
		e.secondBuf = e.secondBuf[:e.MaxLength]
	}

	if areEqual(e.primBuf, e.secondBuf) {
		return string(e.primBuf), ""
	}

	return string(e.primBuf), string(e.secondBuf)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
// Detailed encoder functions
//////////////////////////////////////////////////////////////////////////////////////////////////////

func (e *Encoder) encodeB() {
	if e.encodeSilentB() {
		return
	}

	// "-mb", e.g", "dumb", already skipped over under
	// 'M', altho it should really be handled here...
	e.metaphAddExactApprox("B", "P")

	// skip double B, or BPx where X isn't H
	if e.charNextIs('B') ||
		(e.charNextIs('P') && e.idx+2 < len(e.in) && e.in[e.idx+2] != 'H') {
		e.idx++
	}
}

// Encodes silent 'B' for cases not covered under "-mb-"
func (e *Encoder) encodeSilentB() bool {
	//'debt', 'doubt', 'subtle'
	if e.stringAt(-2, "DEBT", "SUBTL", "SUBTIL") || e.stringAt(-3, "DOUBT") {
		e.metaphAdd('T')
		e.idx++
		return true
	}

	return false
}

func (e *Encoder) encodeC() {
	if e.encodeSilentCAtBeginning() ||
		e.encodeCaToS() ||
		e.encodeCoToS() ||
		e.encodeCh() ||
		e.encodeCcia() ||
		e.encodeCc() ||
		e.encodeCkCgCq() ||
		e.encodeCFrontVowel() ||
		e.encodeSilentC() ||
		e.encodeCz() ||
		e.encodeCs() {
		return
	}

	if !e.stringAt(-1, "C", "K", "G", "Q") {
		e.metaphAdd('K')
	}

	//name sent in 'mac caffrey', 'mac gregor
	if e.stringAt(1, " C", " Q", " G") {
		e.idx++
	} else {
		if e.stringAt(1, "C", "K", "Q") && !e.stringAt(1, "CE", "CI") {
			e.idx++ // increment 1 here, so adjust offsets below
			// account for combinations such as Ro-ckc-liffe
			if e.stringAt(1, "C", "K", "Q") && !e.stringAt(2, "CE", "CI") {
				e.idx++
			}
		}
	}
}

func (e *Encoder) encodeSilentCAtBeginning() bool {
	if e.idx == 0 && e.stringAt(0, "CT", "CN") {
		return true
	}
	return false
}

//Encodes exceptions where "-CA-" should encode to S
//instead of K including cases where the cedilla has not been used
func (e *Encoder) encodeCaToS() bool {
	// Special case: 'caesar'.
	// Also, where cedilla not used, as in "linguica" => LNKS
	if (e.idx == 0 && e.stringAt(0, "CAES", "CAEC", "CAEM")) ||
		e.stringStart("FACADE", "FRANCAIS", "FRANCAIX", "LINGUICA", "GONCALVES", "PROVENCAL") {
		e.metaphAdd('S')
		e.advanceCounter(1, 0)
		return true
	}

	return false
}

//Encodes exceptions where "-CO-" encodes to S instead of K
//including cases where the cedilla has not been used
func (e *Encoder) encodeCoToS() bool {
	// e.g. 'coelecanth' => SLKN0
	if e.stringAt(0, "COEL") && (e.isVowelAt(4) || e.idx+3 == e.lastIdx) ||
		e.stringAt(0, "COENA", "COENO") || e.stringStart("GARCON", "FRANCOIS", "MELANCON") {

		e.metaphAdd('S')
		e.advanceCounter(2, 0)
		return true
	}

	return false
}

func (e *Encoder) encodeCh() bool {
	if !e.stringAt(0, "CH") {
		return false
	}

	if e.encodeChae() ||
		e.encodeChToH() ||
		e.encodeSilentCh() ||
		e.encodeArch() ||
		e.encodeChToX() ||
		e.encodeEnglishChToK() ||
		e.encodeGermanicChToK() ||
		e.encodeGreekChInitial() ||
		e.encodeGreekChNonInitial() {
		return true
	}

	if e.idx > 0 {
		if e.stringStart("MC") && e.idx == 1 {
			//e.g., "McHugh"
			e.metaphAdd('K')
		} else {
			e.metaphAddAlt('X', 'K')
		}
	} else {
		e.metaphAdd('X')
	}

	e.idx++
	return true
}

func (e *Encoder) encodeChae() bool {
	// e.g. 'michael'
	if e.idx > 0 && e.stringAt(2, "AE") {
		if e.stringStart("RACHAEL") {
			e.metaphAdd('X')
		} else if !e.stringAt(-1, "C", "K", "G", "Q") {
			e.metaphAdd('K')
		}

		e.advanceCounter(3, 1)
		return true
	}

	return false
}

// Encodes transliterations from the hebrew where the
// sound 'kh' is represented as "-CH-". The normal pronounciation
// of this in english is either 'h' or 'kh', and alternate
// spellings most often use "-H-"
func (e *Encoder) encodeChToH() bool {
	// hebrew => 'H', e.g. 'channukah', 'chabad'
	if (e.idx == 0 &&
		(e.stringAt(2, "AIM", "ETH", "ELM", "ASID", "AZAN",
			"UPPAH", "UTZPA", "ALLAH", "ALUTZ", "AMETZ",
			"ESHVAN", "ADARIM", "ANUKAH", "ALLLOTH", "ANNUKAH", "AROSETH"))) ||
		e.stringAt(-3, "CLACHAN") {

		e.metaphAdd('H')
		e.advanceCounter(2, 1)
		return true
	}

	return false
}

func (e *Encoder) encodeSilentCh() bool {
	if e.stringAt(-2, "YACHT", "FUCHSIA") ||
		e.stringStart("STRACHAN", "CRICHTON") ||
		(e.stringAt(-3, "DRACHM") && !e.stringAt(-3, "DRACHMA")) {
		e.idx++
		return true
	}

	return false
}

func (e *Encoder) encodeChToX() bool {
	// e.g. 'approach', 'beach'
	if (e.stringAt(-2, "OACH", "EACH", "EECH", "OUCH", "OOCH", "MUCH", "SUCH") && !e.stringAt(-3, "JOACH")) ||
		e.stringAtEnd(-1, "ACHA", "ACHO") || // e.g. 'dacha', 'macho'
		e.stringAtEnd(0, "CHOT", "CHOD", "CHAT") ||
		(e.stringAtEnd(-1, "OCHE") && !e.stringAt(-2, "DOCHE")) ||
		e.stringAt(-4, "ATTACH", "DETACH", "KOVACH", "PARACHUT") ||
		e.stringAt(-5, "SPINACH", "MASSACHU") ||
		e.stringStart("MACHAU") ||
		(e.stringAt(-3, "THACH") && !e.stringAt(2, "E")) || // no ACHE
		e.stringAt(-2, "VACHON") {

		e.metaphAdd('X')
		e.idx++
		return true
	}

	return false
}

func (e *Encoder) encodeEnglishChToK() bool {
	//'ache', 'echo', alternate spelling of 'michael'
	if (e.idx == 1 && rootOrInflections(e.in, "ACHE")) ||
		((e.idx > 3 && rootOrInflections(e.in[e.idx-1:], "ACHE")) &&
			e.stringStart("EAR", "HEAD", "BACK", "HEART", "BELLY", "TOOTH")) ||
		e.stringAt(-1, "ECHO") ||
		e.stringAt(-2, "MICHEAL") ||
		e.stringAt(-4, "JERICHO") ||
		e.stringAt(-5, "LEPRECH") {

		e.metaphAddAlt('K', 'X')
		e.idx++
		return true
	}

	return false
}

func (e *Encoder) encodeGermanicChToK() bool {
	// various germanic
	// "<consonant><vowel>CH-"implies a german word where 'ch' => K

	if (e.idx > 1 &&
		!e.isVowelAt(-2) &&
		e.stringAt(-1, "ACH") &&
		!e.stringAt(-2, "MACHADO", "MACHUCA", "LACHANC", "LACHAPE", "KACHATU") &&
		!e.stringAt(-3, "KHACHAT") &&
		(!e.charAt(2, 'I') && (!e.charAt(2, 'E') || e.stringAt(-2, "BACHER", "MACHER", "MACHEN", "LACHER"))) ||
		// e.g. 'brecht', 'fuchs'
		(e.stringAt(2, "T", "S") && !(e.stringStart("WHICHSOEVER", "LUNCHTIME"))) ||
		// e.g. 'andromache'
		e.stringStart("SCHR") ||
		(e.idx > 2 && e.stringAt(-2, "MACHE")) ||
		(e.idx == 2 && e.stringAt(-2, "ZACH")) ||
		e.stringAt(-4, "SCHACH") ||
		e.stringAt(-1, "ACHEN") ||
		e.stringAt(-3, "SPICH", "ZURCH", "BUECH") ||
		(e.stringAt(-3, "KIRCH", "JOACH", "BLECH", "MALCH") &&
			!(e.stringAt(-3, "KIRCHNER") || e.idx+1 == e.lastIdx)) || // "kirch" and "blech" both get 'X'
		e.stringAtEnd(-2, "NICH", "LICH", "BACH") ||
		(e.stringAtEnd(-3, "URICH", "BRICH", "ERICH", "DRICH", "NRICH") &&
			!e.stringAtEnd(-5, "ALDRICH") &&
			!e.stringAtEnd(-6, "GOODRICH") &&
			!e.stringAtEnd(-7, "GINGERICH"))) ||
		e.stringAtEnd(-4, "ULRICH", "LFRICH", "LLRICH", "EMRICH", "ZURICH", "EYRICH") ||
		// e.g., 'wachtler', 'wechsler', but not 'tichner'
		((e.stringAt(-1, "A", "O", "U", "E") || e.idx == 0) &&
			e.stringAt(2, "L", "R", "N", "M", "B", "H", "F", "V", "W", " ")) {

		// "CHR/L-" e.g. 'chris' do not get
		// alt pronunciation of 'X'
		if e.stringAt(2, "R", "L") || e.isSlavoGermanic() {
			e.metaphAdd('K')
		} else {
			e.metaphAddAlt('K', 'X')
		}
		e.idx++
		return true
	}

	return false
}

// Encode "-ARCH-". Some occurances are from greek roots and therefore encode
// to 'K', others are from english words and therefore encode to 'X'
func (e *Encoder) encodeArch() bool {
	if e.stringAt(-2, "ARCH") {
		// "-ARCH-" has many combining forms where "-CH-" => K because of its
		// derivation from the greek
		if ((e.isVowelAt(2) && e.stringAt(-2, "ARCHA", "ARCHI", "ARCHO", "ARCHU", "ARCHY")) ||
			e.stringAt(-2, "ARCHEA", "ARCHEG", "ARCHEO", "ARCHET", "ARCHEL", "ARCHES", "ARCHEP", "ARCHEM", "ARCHEN") ||
			e.stringAtEnd(-2, "ARCH") ||
			e.stringStart("MENARCH")) &&
			(!rootOrInflections(e.in, "ARCH") &&
				!e.stringAt(-4, "SEARCH", "POARCH") &&
				!e.stringStart("ARCHER", "ARCHIE", "ARCHENEMY", "ARCHIBALD", "ARCHULETA", "ARCHAMBAU") &&
				!((((e.stringAt(-3, "LARCH", "MARCH", "PARCH") ||
					e.stringAt(-4, "STARCH")) &&
					!e.stringStart("EPARCH", "NOMARCH", "EXILARCH", "HIPPARCH", "MARCHESE", "ARISTARCH", "MARCHETTI")) ||
					rootOrInflections(e.in, "STARCH")) &&
					(!e.stringAt(-2, "ARCHU", "ARCHY") || e.stringStart("STARCHY")))) {

			e.metaphAddAlt('K', 'X')
		} else {
			e.metaphAdd('X')
		}
		e.idx++
		return true
	}

	return false
}

func (e *Encoder) encodeGreekChInitial() bool {
	// greek roots e.g. 'chemistry', 'chorus', ch at beginning of root
	if (e.stringAt(0, "CHAMOM", "CHARAC", "CHARIS", "CHARTO", "CHARTU", "CHARYB", "CHRIST", "CHEMIC", "CHILIA") ||
		(e.stringAt(0, "CHEMI", "CHEMO", "CHEMU", "CHEMY", "CHOND", "CHONA", "CHONI", "CHOIR", "CHASM",
			"CHARO", "CHROM", "CHROI", "CHAMA", "CHALC", "CHALD", "CHAET", "CHIRO", "CHILO", "CHELA", "CHOUS",
			"CHEIL", "CHEIR", "CHEIM", "CHITI", "CHEOP") && !(e.stringAt(0, "CHEMIN") || e.stringAt(-2, "ANCHONDO"))) ||
		(e.stringAt(0, "CHISM", "CHELI") &&
			// exclude spanish "machismo"
			!(e.stringStart("MICHEL", "MACHISMO", "RICHELIEU", "REVANCHISM") ||
				e.stringExact("CHISM"))) ||
		// include e.g. "chorus", "chyme", "chaos"
		(e.stringAt(0, "CHOR", "CHOL", "CHYM", "CHYL", "CHLO", "CHOS", "CHUS", "CHOE") && !e.stringStart("CHOLLO", "CHOLLA", "CHORIZ")) ||
		// "chaos" => K but not "chao"
		(e.stringAt(0, "CHAO") && e.idx+3 != e.lastIdx) ||
		// e.g. "abranchiate"
		(e.stringAt(0, "CHIA") && !(e.stringStart("CHIAPAS", "APPALACHIA"))) ||
		// e.g. "chimera"
		e.stringAt(0, "CHIMERA", "CHIMAER", "CHIMERI") ||
		// e.g. "chameleon"
		e.stringStart("CHAME", "CHELO", "CHITO") ||
		// e.g. "spirochete"
		((e.idx+4 == e.lastIdx || e.idx+5 == e.lastIdx) && e.stringAt(-1, "OCHETE"))) &&
		// more exceptions where "-CH-" => X e.g. "chortle", "crocheter"
		!(e.stringExact("CHORE", "CHOLO", "CHOLA") ||
			e.stringAt(0, "CHORT", "CHOSE") ||
			e.stringAt(-3, "CROCHET") ||
			e.stringStart("CHEMISE", "CHARISE", "CHARISS", "CHAROLE")) {

		if e.stringAt(2, "R", "L") {
			e.metaphAdd('K')
		} else {
			e.metaphAddAlt('K', 'X')
		}
		e.idx++
		return true
	}

	return false
}

func (e *Encoder) encodeGreekChNonInitial() bool {
	//greek & other roots e.g. 'tachometer', 'orchid', ch in middle or end of root
	if e.stringAt(-2, "LYCHN", "TACHO", "ORCHO", "ORCHI", "LICHO", "ORCHID", "NICHOL",
		"MECHAN", "LICHEN", "MACHIC", "PACHEL", "RACHIF", "RACHID",
		"RACHIS", "RACHIC", "MICHAL", "ORCHESTR") ||
		e.stringAt(-3, "MELCH", "GLOCH", "TRACH", "TROCH", "BRACH", "SYNCH", "PSYCH",
			"STICH", "PULCH", "EPOCH") ||
		(e.stringAt(-3, "TRICH") && !e.stringAt(-5, "OSTRICH")) ||
		(e.stringAt(-2, "TYCH", "TOCH", "BUCH", "MOCH", "CICH", "DICH", "NUCH", "EICH", "LOCH",
			"DOCH", "ZECH", "WYCH") && !(e.stringAt(-4, "INDOCHINA") || e.stringAt(-2, "BUCHON"))) ||
		((e.idx == 1 || e.idx == 2) && e.stringAt(-1, "OCHER", "ECHIN", "ECHID")) ||
		e.stringAt(-4, "BRONCH", "STOICH", "STRYCH", "TELECH", "PLANCH", "CATECH", "MANICH", "MALACH",
			"BIANCH", "DIDACH", "BRANCHIO", "BRANCHIF") ||
		e.stringStart("ICHA", "ICHN") ||
		(e.stringAt(-1, "ACHAB", "ACHAD", "ACHAN", "ACHAZ") && !e.stringAt(-2, "MACHADO", "LACHANC")) ||
		e.stringAt(-1, "ACHISH", "ACHILL", "ACHAIA", "ACHENE", "ACHAIAN", "ACHATES", "ACHIRAL", "ACHERON",
			"ACHILLEA", "ACHIMAAS", "ACHILARY", "ACHELOUS", "ACHENIAL", "ACHERNAR",
			"ACHALASIA", "ACHILLEAN", "ACHIMENES", "ACHIMELECH", "ACHITOPHEL") ||
		// e.g. 'inchoate'
		(e.idx == 2 && (e.stringStart("INCHOA")) ||
			// e.g. 'ischemia'
			e.stringStart("ISCH")) ||
		// e.g. 'ablimelech', 'antioch', 'pentateuch'
		(e.idx+1 == e.lastIdx && e.stringAt(-1, "A", "O", "U", "E") &&
			!(e.stringStart("DEBAUCH") || e.stringAt(-2, "MUCH", "SUCH", "KOCH") ||
				e.stringAt(-5, "OODRICH", "ALDRICH"))) {

		e.metaphAddAlt('K', 'X')
		e.idx++
		return true
	}

	return false
}

//Encodes reliably italian "-CCIA-"
func (e *Encoder) encodeCcia() bool {
	//e.g., 'focaccia'
	if e.stringAt(1, "CIA") {
		e.metaphAddAlt('X', 'S')
		e.idx++
		return true
	}

	return false
}

func (e *Encoder) encodeCc() bool {
	//double 'C', but not if e.g. 'McClellan'
	if e.stringAt(0, "CC") && !(e.idx == 1 && e.in[0] == 'M') {
		// exception
		if e.stringAt(-3, "FLACCID") {
			e.metaphAdd('S')
			e.advanceCounter(2, 1)
			return true
		}

		//'bacci', 'bertucci', other italian
		if e.stringAtEnd(2, "I") ||
			e.stringAt(2, "IO") || e.stringAtEnd(2, "INO", "INI") {
			e.metaphAdd('X')
			e.advanceCounter(2, 1)
			return true
		}

		//'accident', 'accede' 'succeed'
		if e.stringAt(2, "I", "E", "Y") && //except 'bellocchio','bacchus', 'soccer' get K
			!(e.charAt(2, 'H') || e.stringAt(-2, "SOCCER")) {
			e.metaphAddStr("KS", "KS")
			e.advanceCounter(2, 1)
			return true
		}
		// Pierce's rule
		e.metaphAdd('K')
		e.idx++
		return true
	}

	return false
}

func (e *Encoder) encodeCkCgCq() bool {
	if e.stringAt(0, "CK", "CG", "CQ") {

		// eastern european spelling e.g. 'gorecki' == 'goresky'
		if e.stringAtEnd(0, "CKI", "CKY") && len(e.in) > 6 {
			e.metaphAddStr("K", "SK")
		} else {
			e.metaphAdd('K')
		}
		e.idx++ // skip the C
		// if there's a C[KGQ][KGQ] then skip that second one too
		if e.stringAt(1, "K", "G", "Q") {
			e.idx++
		}

		return true
	}

	return false
}

//Encode cases where "C" preceeds a front vowel such as "E", "I", or "Y".
//These cases most likely => S or X
func (e *Encoder) encodeCFrontVowel() bool {
	if e.stringAt(0, "CI", "CE", "CY") {
		if e.encodeBritishSilentCE() ||
			e.encodeCe() ||
			e.encodeCi() ||
			e.encodeLatinateSuffixes() {

			e.advanceCounter(1, 0)
			return true
		}

		e.metaphAdd('S')
		e.advanceCounter(1, 0)
		return true
	}

	return false
}

func (e *Encoder) encodeBritishSilentCE() bool {
	// english place names like e.g.'gloucester' pronounced glo-ster
	if e.stringAtEnd(1, "ESTER") || e.stringAt(1, "ESTERSHIRE") {
		return true
	}

	return false
}

func (e *Encoder) encodeCe() bool {
	// 'ocean', 'commercial', 'provincial', 'cello', 'fettucini', 'medici'
	if (e.stringAt(1, "EAN") && e.isVowelAt(-1)) ||
		(e.stringAtEnd(-1, "ACEA") && !e.stringStart("PANACEA")) || // e.g. 'rosacea'
		e.stringAt(1, "ELLI", "ERTO", "EORL") || // e.g. 'botticelli', 'concerto'
		e.stringAtEnd(-3, "CROCE") || // some italian names familiar to americans
		e.stringAt(-3, "DOLCE") ||
		e.stringAtEnd(1, "ELLO") { // e.g. cello

		e.metaphAddAlt('X', 'S')
		return true
	}

	return false
}

func (e *Encoder) encodeCi() bool {
	// with consonant before C
	// e.g. 'fettucini', but exception for the americanized pronunciation of 'mancini'

	if (e.stringAtEnd(1, "INI") && !e.stringExact("MANCINI")) ||
		e.stringAtEnd(-1, "ICI") || // e.g. 'medici'
		e.stringAt(-1, "RCIAL", "NCIAL", "RCIAN", "UCIUS") || // e.g. "commercial', 'provincial', 'cistercian'
		e.stringAt(-3, "MARCIA") || // special cases
		e.stringAt(-2, "ANCIENT") {
		e.metaphAddAlt('X', 'S')
		return true
	}

	// exception
	if e.stringAt(-4, "COERCION") {
		e.metaphAdd('J')
		return true
	}

	// with vowel before C (or at beginning?)
	if (e.stringAt(0, "CIO", "CIE", "CIA") && e.isVowelAt(-1)) ||
		e.stringAt(1, "IAO") {

		if (e.stringAt(0, "CIAN", "CIAL", "CIAO", "CIES", "CIOL", "CION") ||
			e.stringAt(-3, "GLACIER") || // exception - "glacier" => 'X' but "spacier" = > 'S'
			e.stringAt(0, "CIENT", "CIENC", "CIOUS", "CIATE", "CIATI", "CIATO", "CIABL", "CIARY") ||
			e.stringAtEnd(0, "CIA", "CIO", "CIAS", "CIOS")) &&
			!(e.stringAt(-4, "ASSOCIATION") || e.stringStart("OCIE") ||
				// exceptions mostly because these names are usually from
				// the spanish rather than the italian in america
				e.stringAt(-2, "LUCIO", "SOCIO", "SOCIE", "MACIAS", "LUCIANO", "HACIENDA") ||
				e.stringAt(-3, "GRACIE", "GRACIA", "MARCIANO") ||
				e.stringAt(-4, "PALACIO", "POLICIES", "FELICIANO") ||
				e.stringAt(-5, "MAURICIO") ||
				e.stringAt(-6, "ANDALUCIA") ||
				e.stringAt(-7, "ENCARNACION")) {

			e.metaphAddAlt('X', 'S')
		} else {
			e.metaphAddAlt('S', 'X')
		}

		return true
	}

	return false
}

func (e *Encoder) encodeLatinateSuffixes() bool {
	if e.stringAt(1, "EOUS", "IOUS") {
		e.metaphAddAlt('X', 'S')
		return true
	}
	return false
}

func (e *Encoder) encodeSilentC() bool {
	if e.stringAt(1, "T", "S") && e.stringStart("INDICT", "TUCSON", "CONNECTICUT") {
		return true
	}

	return false
}

// Encodes slavic spellings or transliterations
// written as "-CZ-"
func (e *Encoder) encodeCz() bool {
	if e.stringAt(1, "Z") && !e.stringAt(-1, "ECZEMA") {
		if e.stringAt(0, "CZAR") {
			e.metaphAdd('S')
		} else {
			// otherwise most likely a czech word...
			e.metaphAdd('X')
		}
		e.idx++
		return true
	}

	return false
}

func (e *Encoder) encodeCs() bool {
	// give an 'etymological' 2nd
	// encoding for "kovacs" so
	// that it matches "kovach"

	if e.stringStart("KOVACS") {
		e.metaphAddStr("KS", "X")
		e.idx++
		return true
	}

	if e.stringAtEnd(-1, "ACS") && !e.stringAt(-4, "ISAACS") {
		e.metaphAdd('X')
		e.idx++
		return true
	}

	return false
}

func (e *Encoder) encodeD() {
	if e.encodeDg() || e.encodeDj() || e.encodeDtDd() ||
		e.encodeDToJ() || e.encodeDous() || e.encodeSilentD() {
		return
	}

	if e.EncodeExact {
		// "final de-voicing" in this case
		// e.g. 'missed' == 'mist'
		if e.stringAtEnd(-3, "SSED") {
			e.metaphAdd('T')
		} else {
			e.metaphAdd('D')
		}
	} else {
		e.metaphAdd('T')
	}
}

func (e *Encoder) encodeDg() bool {
	if e.stringAt(0, "DG") {
		// excludes exceptions e.g. 'edgar',
		// or cases where 'g' is first letter of combining form
		// e.g. 'handgun', 'waldglas'
		if e.stringAt(2, "A", "O") ||
			// e.g. "midgut"
			// e.g. "handgrip"
			// e.g. "mudgard"
			// e.g. "woodgrouse"
			e.stringAt(1, "GUN", "GUT", "GEAR", "GLAS", "GRIP", "GREN", "GILL", "GRAF",
				"GUARD", "GUILT", "GRAVE", "GRASS", "GROUSE") {

			e.metaphAddExactApprox("DG", "TK")
		} else {
			// e.g. "edge", "abridgment"
			e.metaphAdd('J')
		}

		e.idx++
		return true
	}

	return false
}

func (e *Encoder) encodeDj() bool {
	// e.g. "adjacent"
	if e.stringAt(0, "DJ") {
		e.metaphAdd('J')
		e.idx++
		return true
	}
	return false
}

func (e *Encoder) encodeDtDd() bool {
	// eat redundant 'T' or 'D'
	if e.stringAt(0, "DT", "DD") {
		if e.stringAt(0, "DTH") {
			e.metaphAddExactApprox("D0", "T0")
			e.idx += 2
		} else {
			if e.EncodeExact {
				// devoice it
				if e.stringAt(0, "DT") {
					e.metaphAdd('T')
				} else {
					e.metaphAdd('D')
				}
			} else {
				e.metaphAdd('T')
			}
			e.idx++
		}

		return true
	}
	return false
}

func (e *Encoder) encodeDToJ() bool {
	// e.g. "module", "adulate"
	if (e.stringAt(0, "DUL") && e.isVowelAt(-1) && e.isVowelAt(3)) ||
		// e.g. "soldier", "grandeur", "procedure"
		e.stringAtEnd(-1, "LDIER", "NDEUR", "EDURE", "RDURE") ||
		e.stringAt(-3, "CORDIAL") ||
		// e.g. "pendulum", "education"
		// e.g. "individual", "individual", "residuum"
		e.stringAt(-1, "ADUA", "IDUA", "IDUU", "NDULA", "NDULU", "EDUCA") {

		e.metaphAddExactApproxAlt("J", "D", "J", "T")
		e.advanceCounter(1, 0)
		return true
	}
	return false
}

func (e *Encoder) encodeDous() bool {
	// e.g. "assiduous", "arduous"
	if e.stringAt(1, "UOUS") {
		e.metaphAddExactApproxAlt("J", "D", "J", "T")
		e.advanceCounter(3, 0)
		return true
	}
	return false
}

func (e *Encoder) encodeSilentD() bool {
	// silent 'D' e.g. 'wednesday', 'handsome'
	return e.stringAt(-2, "WEDNESDAY") ||
		e.stringAt(-3, "HANDKER", "HANDSOM", "WINDSOR") ||
		// french silent D at end in words or names familiar to americans
		e.stringEnd("PERNOD", "ARTAUD", "RENAUD", "RIMBAUD", "MICHAUD", "BICHAUD")
}

func (e *Encoder) encodeF() {
	// Encode cases where "-FT-" => "T" is usually silent
	// e.g. 'often', 'soften'
	// This should really be covered under "T"!
	if e.stringAt(-1, "OFTEN") {
		e.metaphAddStr("F", "FT")
		e.idx++
		return
	}

	// eat redundant 'F'
	if e.charNextIs('F') {
		e.idx++
	}
	e.metaphAdd('F')
}

//800
func (e *Encoder) encodeG() {
	//todo: special cases

	if !e.stringAt(-1, "C", "K", "G", "Q") {
		e.metaphAddExactApprox("G", "K")
	}
}

func (e *Encoder) encodeH() {
	if e.encodeInitialSilentH() || e.encodeInitialHs() ||
		e.encodeInitialHuHw() || e.encodeNonInitialSilentH() {
		return
	}

	// only keep if first & before vowel or btw. 2 vowels
	if !e.encodeHPronounced() {
		//e.idx++ ?
	}
}

func (e *Encoder) encodeInitialSilentH() bool {
	// 'hour', 'herb', 'heir', 'honor'
	if e.stringAt(1, "OUR", "ERB", "EIR", "ONOR", "ONOUR", "ONEST") {
		// british pronounce H in this word
		// americans give it 'H' for the name,
		// no 'H' for the plant
		if e.stringAtStart(0, "HERB") {
			if e.EncodeVowels {
				e.metaphAddStr("HA", "A")
			} else {
				e.metaphAddAlt('H', 'A')
			}
		} else if e.idx == 0 || e.EncodeVowels {
			e.metaphAdd('A')
		}

		// don't encode vowels twice
		e.idx = e.skipVowels(e.idx + 1)
		return true
	}

	return false
}

func (e *Encoder) encodeInitialHs() bool {
	// old chinese pinyin transliteration
	// e.g., 'HSIAO'
	if e.stringAtStart(0, "HS") {
		e.metaphAdd('X')
		e.idx++
		return true
	}
	return false
}

func (e *Encoder) encodeInitialHuHw() bool {
	// spanish spellings and chinese pinyin transliteration
	if e.stringStart("HUA", "HUE", "HWA") && !e.stringAt(0, "HUEY") {
		e.metaphAdd('A')

		if !e.EncodeVowels {
			e.idx += 2
		} else {
			e.idx++
			// don't encode vowels twice
			//TODO: why not this: e.idx = e.skipVowels(e.idx + 1)
			for e.isVowelAt(0) || e.charAt(0, 'W') {
				e.idx++
			}
			e.idx-- // give back one that's going to be added in the main loop
		}
		return true
	}

	return false
}

func (e *Encoder) encodeNonInitialSilentH() bool {
	if e.stringAt(-2, "NIHIL", "VEHEM", "LOHEN", "NEHEM", "MAHON", "MAHAN", "COHEN", "GAHAN") ||
		e.stringAt(-3, "TOUHY", "GRAHAM", "PROHIB", "FRAHER", "TOOHEY", "TOUHEY") ||
		e.stringStart("CHIHUAHUA") {
		if e.EncodeVowels {
			e.idx++
		} else {
			e.idx = e.skipVowels(e.idx + 1)
		}
		return true
	}
	return false
}

func (e *Encoder) encodeHPronounced() bool {
	if ((e.idx == 0 || e.isVowelAt(-1) || (e.idx > 0 && e.charAt(-1, 'W'))) &&
		e.isVowelAt(1)) ||
		// e.g. 'alWahhab'
		(e.charNextIs('H') && e.isVowelAt(2)) {

		e.metaphAdd('H')
		e.advanceCounter(1, 0)
		return true
	}

	return false
}

func (e *Encoder) encodeJ() {
	if e.encodeSpanishJ() || e.encodeSpanishOjUj() {
		return
	}

	//e.encodeOtherJ()
	if e.idx == 0 {
		if e.encodeGermanJ() {
			return
		} else if e.encodeJToJ() {
			return
		}
	} else {
		if e.encodeSpanishJ2() {
			return
		} else if !e.encodeJAsVowel() {
			e.metaphAdd('J')
		}

		// it could happen! e.g. "hajj"
		// eat redundant 'J'
		if e.charNextIs('J') {
			e.idx++
		}
	}
}

func (e *Encoder) encodeSpanishJ() bool {
	//obvious spanish, e.g. "jose", "san jacinto"
	if (e.stringAt(1, "UAN", "ACI", "ALI", "EFE", "ICA", "IME", "OAQ", "UAR") &&
		!e.stringAt(0, "JIMERSON", "JIMERSEN")) ||
		e.stringAtEnd(1, "OSE") ||
		e.stringAt(1, "EREZ", "UNTA", "AIME", "AVIE", "AVIA", "IMINEZ", "ARAMIL") ||
		e.stringAtEnd(-2, "MEJIA") ||
		e.stringAt(-2, "TEJED", "TEJAD", "LUJAN", "FAJAR", "BEJAR", "BOJOR", "CAJIG",
			"DEJAS", "DUJAR", "DUJAN", "MIJAR", "MEJOR", "NAJAR",
			"NOJOS", "RAJED", "RIJAL", "REJON", "TEJAN", "UIJAN") ||
		e.stringAt(-3, "ALEJANDR", "GUAJARDO", "TRUJILLO") ||
		(e.stringAt(-2, "RAJAS") && e.idx > 2) ||
		(e.stringAt(-2, "MEJIA") && !e.stringAt(-2, "MEJIAN")) ||
		e.stringAt(-1, "OJEDA") ||
		e.stringAt(-3, "LEIJA", "MINJA", "VIAJES", "GRAJAL") ||
		e.stringAt(0, "JAUREGUI") ||
		e.stringAt(-4, "HINOJOSA") ||
		e.stringStart("SAN ") ||
		((e.idx+1 == e.lastIdx) && e.charAt(1, 'O') && !e.stringStart("TOJO", "BANJO", "MARYJO")) {

		// americans pronounce "juan" as 'wan'
		// and "marijuana" and "tijuana" also
		// do not get the 'H' as in spanish, so
		// just treat it like a vowel in these cases

		if !(e.stringAt(0, "JUAN") || e.stringAt(0, "JOAQ")) {
			e.metaphAdd('H')
		} else if e.idx == 0 {
			e.metaphAdd('A')
		}
		e.advanceCounter(1, 0)
		return true
	}

	// Jorge gets 2nd HARHA. also JULIO, JESUS
	if e.stringAt(1, "ORGE", "ULIO", "ESUS") && !e.stringStart("JORGEN") {
		// get both consonants for "jorge"
		if e.stringAtEnd(1, "ORGE") {
			if e.EncodeVowels {
				e.metaphAddStr("JARJ", "HARHA")
			} else {
				e.metaphAddStr("JRJ", "HRH")
			}
			e.advanceCounter(4, 4)
			return true
		}
		e.metaphAddAlt('J', 'H')
		e.advanceCounter(1, 0)
		return true
	}

	return false
}

func (e *Encoder) encodeGermanJ() bool {
	if e.stringAt(1, "AH", "UGO") || e.stringExact("JOHANN") ||
		(e.stringAt(1, "UNG") && !e.charAt(4, 'L')) {

		e.metaphAdd('A')
		e.advanceCounter(1, 0)
		return true
	}

	return false
}

func (e *Encoder) encodeSpanishOjUj() bool {
	if e.stringAt(1, "OJOBA", "UJUY") {
		if e.EncodeVowels {
			e.metaphAddStr("HAH", "HAH")
		} else {
			e.metaphAddStr("HH", "HH")
		}

		e.advanceCounter(3, 2)
		return true
	}

	return false
}

func (e *Encoder) encodeJToJ() bool {
	if e.isVowelAt(1) {
		if e.idx == 0 && e.namesBeginningWithJThatGetAltY() {
			// 'Y' is a vowel so encode
			// is as 'A'
			if e.EncodeVowels {
				e.metaphAddStr("JA", "A")
			} else {
				e.metaphAddAlt('J', 'A')
			}
		} else {
			if e.EncodeVowels {
				e.metaphAddStr("JA", "JA")
			} else {
				e.metaphAdd('J')
			}
		}
		e.idx = e.skipVowels(e.idx + 1)
		return false
	}

	e.metaphAdd('J')
	return true
}

func (e *Encoder) encodeSpanishJ2() bool {
	// spanish forms e.g. "brujo", "badajoz"
	if e.stringAtStart(-2, "BOJA", "BAJA", "BEJA", "BOJO", "MOJA", "MOJI", "MEJI") ||
		e.stringAtStart(-3, "FRIJO", "BRUJO", "BRUJA", "GRAJE", "GRIJA", "LEIJA", "QUIJA") ||
		(e.stringAtEnd(-1, "OJA", "EJA", "AJOS", "EJOS", "OJAS", "OJOS", "UJON",
			"AJOZ", "AJAL", "UJAR", "EJON", "EJAN", "AJARA") && !e.stringStart("DEJA")) {

		e.metaphAdd('H')
		e.advanceCounter(1, 0)
		return true
	}

	return false
}

func (e *Encoder) encodeJAsVowel() bool {
	if e.stringAt(0, "JEWSK") {
		e.metaphAddAlt('J', unicode.ReplacementChar)
		return true
	}

	// e.g. "stijl", "sejm" - dutch, scandanavian, and eastern european spellings
	// except words from hindi and arabic
	if (e.stringAt(1, "L", "T", "K", "S", "N", "M") && !e.stringAt(2, "A")) ||
		e.stringStart("FJ", "WOJ", "LJUB", "BJOR", "HAJEK", "HALLELUJA", "LJUBLJANA") ||
		// e.g. 'rekjavik', 'blagojevic'
		e.stringAt(0, "JAVIK", "JEVIC") ||
		e.stringExact("SONJA", "TANJA", "TONJA") {

		return true
	}

	return false
}

func (e *Encoder) namesBeginningWithJThatGetAltY() bool {
	return e.stringStart("JAN", "JON", "JAN", "JIN", "JEN",
		"JUHL", "JULY", "JOEL", "JOHN", "JOSH", "JUDE", "JUNE", "JONI", "JULI", "JENA",
		"JUNG", "JINA", "JANA", "JENI", "JOEL", "JANN", "JONA", "JENE", "JULE", "JANI", "JONG", "JOHN",
		"JEAN", "JUNG", "JONE", "JARA", "JUST", "JOST", "JAHN", "JACO", "JANG", "JUDE", "JONE",
		"JOANN", "JANEY", "JANAE", "JOANA", "JUTTA", "JULEE", "JANAY", "JANEE", "JETTA",
		"JOHNA", "JOANE", "JAYNA", "JANES", "JONAS", "JONIE", "JUSTA", "JUNIE", "JUNKO", "JENAE",
		"JULIO", "JINNY", "JOHNS", "JACOB", "JETER", "JAFFE", "JESKE", "JANKE", "JAGER", "JANIK",
		"JANDA", "JOSHI", "JULES", "JANTZ", "JEANS", "JUDAH", "JANUS", "JENNY", "JENEE", "JONAH",
		"JONAS", "JACOB", "JOSUE", "JOSEF", "JULES", "JULIE", "JULIA", "JANIE", "JANIS", "JENNA",
		"JANNA", "JEANA", "JENNI", "JEANE", "JONNA",
		"JORDAN", "JORDON", "JOSEPH", "JOSHUA", "JOSIAH", "JOSPEH", "JUDSON", "JULIAN",
		"JULIUS", "JUNIOR", "JUDITH", "JOESPH", "JOHNIE", "JOANNE", "JEANNE", "JOANNA", "JOSEFA",
		"JULIET", "JANNIE", "JANELL", "JASMIN", "JANINE", "JOHNNY", "JEANIE", "JEANNA", "JOHNNA",
		"JOELLE", "JOVITA", "JOSEPH", "JONNIE", "JANEEN", "JANINA", "JOANIE", "JAZMIN", "JOHNIE",
		"JANENE", "JOHNNY", "JONELL", "JENELL", "JANETT", "JANETH", "JENINE", "JOELLA", "JOEANN",
		"JULIAN", "JOHANA", "JENICE", "JANNET", "JANISE", "JULENE", "JOSHUA", "JANEAN", "JAIMEE",
		"JOETTE", "JANYCE", "JENEVA", "JORDAN", "JACOBS", "JENSEN", "JOSEPH", "JANSEN", "JORDON",
		"JULIAN", "JAEGER", "JACOBY", "JENSON", "JARMAN", "JOSLIN", "JESSEN", "JAHNKE", "JACOBO",
		"JULIEN", "JOSHUA", "JEPSON", "JULIUS", "JANSON", "JACOBI", "JUDSON", "JARBOE", "JOHSON",
		"JANZEN", "JETTON", "JUNKER", "JONSON", "JAROSZ", "JENNER", "JAGGER", "JASMIN", "JEPSEN",
		"JORDEN", "JANNEY", "JUHASZ", "JERGEN", "JAKOB",
		"JOHNSON", "JOHNNIE", "JASMINE", "JEANNIE", "JOHANNA", "JANELLE", "JANETTE",
		"JULIANA", "JUSTINA", "JOSETTE", "JOELLEN", "JENELLE", "JULIETA", "JULIANN", "JULISSA",
		"JENETTE", "JANETTA", "JOSELYN", "JONELLE", "JESENIA", "JANESSA", "JAZMINE", "JEANENE",
		"JOANNIE", "JADWIGA", "JOLANDA", "JULIANE", "JANUARY", "JEANICE", "JANELLA", "JEANETT",
		"JENNINE", "JOHANNE", "JOHNSIE", "JANIECE", "JOHNSON", "JENNELL", "JAMISON", "JANSSEN",
		"JOHNSEN", "JARDINE", "JAGGERS", "JURGENS", "JOURDAN", "JULIANO", "JOSEPHS", "JHONSON",
		"JOZWIAK", "JANICKI", "JELINEK", "JANSSON", "JOACHIM", "JANELLE", "JACOBUS", "JENNING",
		"JANTZEN", "JOHNNIE",
		"JOSEFINA", "JEANNINE", "JULIANNE", "JULIANNA", "JONATHAN", "JONATHON", "JEANETTE",
		"JANNETTE", "JEANETTA", "JOHNETTA", "JENNEFER", "JULIENNE", "JOSPHINE", "JEANELLE", "JOHNETTE",
		"JULIEANN", "JOSEFINE", "JULIETTA", "JOHNSTON", "JACOBSON", "JACOBSEN", "JOHANSEN", "JOHANSON",
		"JAWORSKI", "JENNETTE", "JELLISON", "JOHANNES", "JASINSKI", "JUERGENS", "JARNAGIN", "JEREMIAH",
		"JEPPESEN", "JARNIGAN", "JANOUSEK",
		"JOHNATHAN", "JOHNATHON", "JORGENSEN", "JEANMARIE", "JOSEPHINA", "JEANNETTE",
		"JOSEPHINE", "JEANNETTA", "JORGENSON", "JANKOWSKI", "JOHNSTONE", "JABLONSKI", "JOSEPHSON",
		"JOHANNSEN", "JURGENSEN", "JIMMERSON", "JOHANSSON",
		"JAKUBOWSKI")
}

func (e *Encoder) encodeK() {
	if !e.encodeSilentK() {
		e.metaphAdd('K')

		// eat redundant K's and Q's
		if e.charAt(1, 'K') || e.charAt(1, 'Q') {
			e.idx++
		}
	}
}

func (e *Encoder) encodeSilentK() bool {
	if e.idx == 0 && e.stringStart("KN") {
		if !e.stringAt(2, "ISH", "ESSET", "IEVEL") {
			return true
		}
	}

	// e.g. "know", "knit", "knob"
	if (e.stringAt(1, "NOW", "NIT", "NOT", "NOB") && !e.stringStart("BANKNOTE")) ||
		e.stringAt(1, "NOCK", "NUCK", "NIFE", "NACK", "NIGHT") {
		// N already encoded before
		// e.g. "penknife"
		if e.idx > 0 && e.charAt(-1, 'N') {
			e.idx++
		}

		return true
	}

	return false
}

func (e *Encoder) encodeL() {
	// logic below needs to know this
	// after 'm_current' variable changed
	saveIdx := e.idx

	e.interpolateVowelWhenConsLAtEnd()

	if e.encodeLelyToL() || e.encodeColonel() || e.encodeFrenchAult() || e.encodeFrenchEuil() ||
		e.encodeFrenchOulx() || e.encodeSilentLInLm() || e.encodeSilentLInLkLv() ||
		e.encodeSilentLInOuld() {
		return
	}

	if e.encodeLlAsVowelCases() {
		return
	}

	e.encodeLeCases(saveIdx)
}

//Cases where an L follows D, G, or T at the end have a schwa pronounced before
//the L
func (e *Encoder) interpolateVowelWhenConsLAtEnd() {
	// e.g. "ertl", "vogl"
	if e.EncodeVowels && e.stringAtEnd(-1, "DL", "GL", "TL") {
		e.metaphAdd('A')
	}
}

func (e *Encoder) encodeLelyToL() bool {
	// e.g. "agilely", "docilely"
	if e.stringAtEnd(-1, "ILELY") {
		e.metaphAdd('L')
		e.idx += 2
		return true
	}
	return false
}

func (e *Encoder) encodeColonel() bool {
	if e.stringAt(-2, "COLONEL") {
		e.metaphAdd('R')
		e.idx++
		return true
	}
	return false
}

func (e *Encoder) encodeFrenchAult() bool {
	// e.g. "renault" and "foucault", well known to americans, but not "fault"
	if e.idx > 3 &&
		(e.stringAt(-3, "RAULT", "NAULT", "BAULT", "SAULT", "GAULT", "CAULT") || e.stringAt(-4, "REAULT", "RIAULT", "NEAULT", "BEAULT")) &&
		!(rootOrInflections(e.in, "ASSAULT") || e.stringAt(-8, "SOMERSAULT") || e.stringAt(-9, "SUMMERSAULT")) {

		e.idx++
		return true
	}

	return false
}

func (e *Encoder) encodeFrenchEuil() bool {
	// e.g. "auteuil"
	if e.stringAtEnd(-3, "EUIL") {
		return true
	}
	return false
}

func (e *Encoder) encodeFrenchOulx() bool {
	// e.g. "proulx"
	if e.stringAtEnd(-2, "OULX") {
		e.idx++
		return true
	}
	return false
}

func (e *Encoder) encodeSilentLInLm() bool {
	if e.stringAt(0, "LM", "LN") {
		// e.g. "lincoln", "holmes", "psalm", "salmon"
		if (e.stringAt(-2, "COLN", "CALM", "BALM", "MALM", "PALM") ||
			e.stringAtEnd(-1, "OLM") ||
			e.stringAt(-3, "PSALM", "QUALM") ||
			e.stringAt(-2, "SALMON", "HOLMES") ||
			e.stringAt(-1, "ALMOND") ||
			e.stringAtStart(-1, "ALMS")) &&
			(!e.stringAt(2, "A") &&
				!e.stringAt(-2, "BALMO", "PALMER", "PALMOR", "BALMER") &&
				!e.stringAt(-3, "THALM")) {
			//noop
		} else {
			e.metaphAdd('L')
		}

		return true
	}

	return false
}

func (e *Encoder) encodeSilentLInLkLv() bool {
	if (e.stringAt(-2, "WALK", "YOLK", "FOLK", "HALF", "TALK", "CALF", "BALK", "CALK") ||
		(e.stringAt(-2, "POLK", "HALV", "SALVE", "CALVE", "SOLDER") && !e.stringAt(-2, "POLKA", "PALKO", "HALVA", "HALVO", "SALVER", "CALVER")) ||
		(e.stringAt(-3, "CAULK", "CHALK", "BAULK", "FAULK") && !e.stringAt(-4, "SCHALK"))) &&
		!e.stringAt(-5, "GONSALVES", "GONCALVES") &&
		!e.stringAt(-2, "BALKAN", "TALKAL") &&
		!e.stringAt(-3, "PAULK", "CHALF") {

		return true
	}

	return false
}

func (e *Encoder) encodeSilentLInOuld() bool {
	// 'would', 'could'
	if e.stringAt(-3, "WOULD", "COULD") ||
		(e.stringAt(-4, "SHOULD") && !e.stringAt(-4, "SHOULDER")) {
		e.metaphAddExactApprox("D", "T")
		e.idx++
		return true
	}
	return false
}

//Encode "-ILLA-" and "-ILLE-" in spanish and french contexts were americans
//know to pronounce it as a 'Y'
func (e *Encoder) encodeLlAsVowelSpecialCases() bool {
	if e.stringAt(-5, "TORTILLA") || e.stringAt(-8, "RATATOUILLE") ||
		// e.g. 'guillermo', "veillard"
		(e.stringStart("GUILL", "VEILL", "GAILL") &&
			// 'guillotine' usually has '-ll-' pronounced as 'L' in english
			!(e.stringAt(-3, "GUILLOT", "GUILLOR", "GUILLEN") || e.stringExact("GUILL"))) ||
		// e.g. "brouillard", "gremillion"
		e.stringStart("ROBILL", "BROUILL", "GREMILL") ||
		// e.g. 'mireille'
		// exception "reveille" usually pronounced as 're-vil-lee'
		(e.stringAtEnd(-2, "EILLE") && !e.stringAt(-5, "REVEILLE")) {

		e.idx++
		return true
	}

	return false
}

func (e *Encoder) encodeLlAsVowel() bool {
	// spanish e.g. "cabrillo", "gallegos" but also "gorilla", "ballerina" -
	// give both pronounciations since an american might pronounce "cabrillo"
	// in the spanish or the american fashion.
	if e.stringAtEnd(-1, "ILLO", "ILLA", "ALLE") ||
		(e.stringEnd("A", "O", "AS", "OS") && e.stringAt(-1, "AL", "IL") && !e.stringAt(-1, "ALLA")) ||
		e.stringStart("LLA", "VILLE", "VILLA", "GALLARDO", "VALLADAR", "MAGALLAN", "CAVALLAR", "BALLASTE") {

		e.metaphAddAlt('L', unicode.ReplacementChar)
		e.idx++
		return true
	}
	return false
}

func (e *Encoder) encodeLlAsVowelCases() bool {
	if e.charNextIs('L') {
		if e.encodeLlAsVowelSpecialCases() {
			return true
		} else if e.encodeLlAsVowel() {
			return true
		}
		e.idx++
	}

	return false
}

func (e *Encoder) encodeVowelLeTransposition(idx int) bool {
	// transposition of vowel sound and L occurs in many words,
	// e.g. "bristle", "dazzle", "goggle" => KAKAL
	offset := e.idx - idx
	if e.EncodeVowels && idx > 1 && !e.isVowelAt(offset-1) && e.charAt(offset+1, 'E') &&
		!e.charAt(offset-1, 'L') && !e.charAt(offset-1, 'R') &&
		// lots of exceptions to this:
		!e.isVowelAt(offset+2) &&
		!e.stringStart("MCCLE", "MCLEL", "EMBLEM", "KADLEC", "ECCLESI", "COMPLEC", "COMPLEJ", "ROBLEDO") &&
		!(idx+2 == e.lastIdx && e.stringAt(offset, "LET")) &&
		!e.stringAt(offset, "LEG", "LER", "LEX", "LESS", "LESQ", "LECT", "LEDG", "LETE", "LETH", "LETS", "LETT",
			"LETUS", "LETIV", "LETELY", "LETTER", "LETION", "LETIAN", "LETING", "LETORY", "LETTING") &&
		// e.g. "complement" !=> KAMPALMENT
		!(e.stringAt(offset, "LEMENT") &&
			!(e.stringAt(-5, "BATTLE", "TANGLE", "PUZZLE", "RABBLE", "BABBLE") || e.stringAt(-4, "TABLE"))) &&
		!(idx+2 == e.lastIdx && e.stringAt(offset-2, "OCLES", "ACLES", "AKLES")) &&
		!e.stringAt(offset-3, "LISLE", "AISLE") && !e.stringStart("ISLE") &&
		!e.stringStart("ROBLES") &&
		!e.stringAt(offset-4, "PROBLEM", "RESPLEN") &&
		!e.stringAt(offset-3, "REPLEN") &&
		!e.stringAt(offset-2, "SPLE") &&
		!e.charAt(offset-1, 'H') && !e.charAt(offset-1, 'W') {

		e.metaphAddStr("AL", "AL")
		e.flagAlInversion = true

		// eat redundant 'L'
		if e.charAt(offset+2, 'L') {
			e.idx = idx + 2
		}
		return true
	}

	return false
}

func (e *Encoder) encodeVowelPreserveVowelAfterL(idx int) bool {
	offset := idx - e.idx

	if e.EncodeVowels && !e.isVowelAt(offset-1) && e.charAt(offset+1, 'E') && idx > 1 &&
		idx+1 != e.lastIdx &&
		!(e.stringAt(offset+1, "ES", "ED") && idx+2 == e.lastIdx) &&
		!e.stringAt(offset-1, "RLEST") {

		e.metaphAddStr("LA", "LA")
		e.idx = e.skipVowels(e.idx + 1)
		return true
	}

	return false
}

func (e *Encoder) encodeLeCases(idx int) {
	if e.encodeVowelLeTransposition(idx) {
		return
	}

	if e.encodeVowelPreserveVowelAfterL(idx) {
		return
	}
	e.metaphAdd('L')
}

func (e *Encoder) encodeM() {
	if e.encodeSilentMAtBeginning() || e.encodeMrAndMrs() ||
		e.encodeMac() || e.encodeMpt() {
		return
	}

	// Silent 'B' should really be handled
	// under 'B", not here under 'M'!
	e.encodeMb()

	e.metaphAdd('M')
}

func (e *Encoder) encodeSilentMAtBeginning() bool {
	return e.stringAtStart(0, "MN")
}

func (e *Encoder) encodeMrAndMrs() bool {
	if e.stringExact("MR") {
		if e.EncodeVowels {
			e.metaphAddStr("MASTAR", "MASTAR")
		} else {
			e.metaphAddStr("MSTR", "MSTR")
		}
		e.idx++
		return true
	} else if e.stringExact("MRS") {
		if e.EncodeVowels {
			e.metaphAddStr("MASAS", "MASAS")
		} else {
			e.metaphAddStr("MSS", "MSS")
		}
		e.idx += 2
		return true
	}

	return false
}

func (e *Encoder) encodeMac() bool {
	// should only find irish and
	// scottish names e.g. 'macintosh'
	if e.stringAtStart(0, "MC", "MACIVER", "MACEWEN", "MACELROY", "MACILROY", "MACINTOSH") {
		if e.EncodeVowels {
			e.metaphAddStr("MAK", "MAK")
		} else {
			e.metaphAddStr("MK", "MK")
		}

		if e.stringStart("MC") {
			// watch out for e.g. "McGeorge"
			if e.stringAt(2, "K", "G", "Q") && !e.stringAt(2, "GEOR") {
				e.idx += 2
			} else {
				e.idx++
			}
		} else {
			e.idx += 2
		}

		return true
	}

	return false
}

func (e *Encoder) encodeMpt() bool {
	if e.stringAt(-2, "COMPTROL") || e.stringAt(-4, "ACCOMPT") {
		e.metaphAdd('N')
		e.idx++
		return true
	}

	return false
}

func (e *Encoder) testSilentMb1() bool {
	// e.g. "LAMB", "COMB", "LIMB", "DUMB", "BOMB"
	// Handle combining roots first
	return e.stringAtStart(-3, "THUMB") || e.stringAtStart(-2, "DUMB", "BOMB", "DAMN", "LAMB", "NUMB", "TOMB")
}

func (e *Encoder) testPronouncedMb() bool {
	return e.stringAt(-2, "NUMBER") ||
		(e.stringAt(2, "A", "O") && !e.stringAt(-2, "DUMBASS")) ||
		e.stringAt(-2, "LAMBEN", "LAMBER", "LAMBET", "TOMBIG", "LAMBRE")
}

func (e *Encoder) testSilentMb2() bool {
	// 'M' is the current letter
	return e.charNextIs('B') && e.idx > 1 &&
		(e.idx+1 == e.lastIdx ||
			// other situations where "-MB-" is at end of root
			// but not at end of word. The tests are for standard
			// noun suffixes.
			// e.g. "climbing" => KLMNK
			e.stringAt(2, "ING", "ABL", "LIKE") ||
			e.stringAtEnd(2, "S") ||
			e.stringAt(-5, "BUNCOMB") ||
			//e.g. "bomber"
			(e.stringAtEnd(2, "ED", "ER") &&
				(e.stringStart("CLIMB", "PLUMB") || !e.stringAt(-1, "IMBER", "AMBER", "EMBER", "UMBER")) &&
				!e.stringAt(-2, "CUMBER", "SOMBER")))
}

func (e *Encoder) testPronouncedMb2() bool {
	// e.g. "bombastic", "umbrage", "flamboyant"
	return e.stringAt(-1, "OMBAS", "OMBAD", "UMBRA") || e.stringAt(-3, "FLAM")
}

func (e *Encoder) testMn() bool {
	return e.charNextIs('N') && (e.idx+1 == e.lastIdx ||
		// or at the end of a word but followed by suffixes
		e.stringAtEnd(2, "S", "LY", "ER", "ED", "ING", "EST") ||
		e.stringAt(-2, "DAMNEDEST") ||
		e.stringAt(-5, "GODDAMNIT"))
}

func (e *Encoder) encodeMb() {
	if e.testSilentMb1() {
		if !e.testPronouncedMb() {
			e.idx++
		}
	} else if e.testSilentMb2() {
		if !e.testPronouncedMb2() {
			e.idx++
		}
	} else if e.testMn() || e.charNextIs('M') {
		e.idx++
	}
}

func (e *Encoder) encodeN() {
	if e.encodeNce() {
		return
	}

	//eat redundant 'N'
	if e.charNextIs('N') {
		e.idx++
	}

	// e.g. "aloneness",
	if !e.stringAt(-3, "MONSIEUR") && !e.stringAt(-3, "NENESS") {
		e.metaphAdd('N')
	}
}

//Encode "-NCE-" and "-NSE-" "entrance" is pronounced exactly the same as
//"entrants"
func (e *Encoder) encodeNce() bool {
	// 'acceptance', 'accountancy'
	if e.stringAt(1, "C", "S") && e.stringAt(2, "E", "Y", "I") &&
		(e.idx+2 == e.lastIdx || (e.idx+3 == e.lastIdx && e.charAt(3, 'S'))) {

		e.metaphAddStr("NTS", "NTS")
		e.idx++
		return true
	}

	return false
}

func (e *Encoder) encodeP() {
	if e.encodeSilentPAtBeginning() || e.encodePt() || e.encodePh() ||
		e.encodePph() || e.encodeRps() || e.encodeCoup() ||
		e.encodePneum() || e.encodePsych() || e.encodePsalm() {
		return
	}

	e.encodePb()

	e.metaphAdd('P')
}

func (e *Encoder) encodeSilentPAtBeginning() bool {
	return e.stringAtStart(0, "PN", "PF", "PS", "PT")
}

func (e *Encoder) encodePt() bool {
	// 'pterodactyl', 'receipt', 'asymptote'
	if e.charNextIs('T') &&
		(e.stringAtStart(0, "PTERO") || e.stringAt(-5, "RECEIPT") || e.stringAt(-4, "ASYMPTOT")) {

		e.metaphAdd('T')
		e.idx++
		return true
	}

	return false
}

//Encode "-PH-", usually as F, with exceptions for cases where it is silent, or
//where the 'P' and 'T' are pronounced seperately because they belong to two
//different words in a combining form
func (e *Encoder) encodePh() bool {
	if e.charNextIs('H') {
		// 'PH' silent in these contexts
		if e.stringAt(0, "PHTHALEIN") || e.stringAtStart(0, "PHTH") || e.stringAt(-3, "APOPHTHEGM") {
			e.metaphAdd('0')
			e.idx += 3
		} else if e.idx > 0 &&
			(e.stringAt(2, "AM", "EAD", "OLE", "ELD", "ILL", "OLD", "EAP", "ERD", "ARD", "ANG",
				"ORN", "EAV", "ART", "OUSE", "AMMER", "AZARD", "UGGER", "OLSTER") && !e.stringAt(-1, "LPHAM")) &&
			!e.stringAt(-3, "LYMPH", "NYMPH") {
			// combining forms
			// 'sheepherd', 'upheaval', 'cupholder'
			e.metaphAdd('P')
			e.advanceCounter(2, 1)
		} else {
			e.metaphAdd('F')
			e.idx++
		}

		return true
	}

	return false
}

func (e *Encoder) encodePph() bool {
	// 'sappho'
	if e.charNextIs('P') && e.idx+2 < len(e.in) && e.charAt(2, 'H') {
		e.metaphAdd('F')
		e.idx += 2
		return true
	}
	return false
}

func (e *Encoder) encodeRps() bool {
	// '-corps-', 'corpsman'
	if e.stringAt(-3, "CORPS") && !e.stringAt(-3, "CORPSE") {
		e.idx++
		return true
	}
	return false
}

func (e *Encoder) encodeCoup() bool {
	// 'coup'
	return e.stringAtEnd(-3, "COUP") && !e.stringAt(-5, "RECOUP")
}

func (e *Encoder) encodePneum() bool {
	// '-pneum-'
	if e.stringAt(1, "NEUM") {
		e.metaphAdd('N')
		e.idx++
		return true
	}
	return false
}

func (e *Encoder) encodePsych() bool {
	// '-psych-'
	if e.stringAt(1, "SYCH") {
		if e.EncodeVowels {
			e.metaphAddStr("SAK", "SAK")
		} else {
			e.metaphAddStr("SK", "SK")
		}
		e.idx += 4
		return true
	}
	return false
}

func (e *Encoder) encodePsalm() bool {
	if e.stringAt(1, "SALM") {
		if e.EncodeVowels {
			e.metaphAddStr("SAM", "SAM")
		} else {
			e.metaphAddStr("SM", "SM")
		}
		e.idx += 4
		return true
	}
	return false
}

func (e *Encoder) encodePb() {
	// e.g. "campbell", "raspberry"
	// eat redundant 'P' or 'B'
	if e.stringAt(1, "P", "B") {
		e.idx++
	}
}

func (e *Encoder) encodeQ() {
	// current pinyin
	if e.stringAt(0, "QIN") {
		e.metaphAdd('X')
		return
	}

	// eat redundant 'Q'
	if e.charNextIs('Q') {
		e.idx++
	}
	e.metaphAdd('K')
}

func (e *Encoder) encodeR() {
	if e.encodeRz() {
		return
	}

	if !e.testSilentR() && !e.encodeVowelReTransposition() {
		e.metaphAdd('R')
	}

	// eat redundant 'R'; also skip 'S' as well as 'R' in "poitiers"
	if e.charNextIs('R') || e.stringAt(-6, "POITIERS") {
		e.idx++
	}
}

//Encode "-RZ-" according to american and polish pronunciations
func (e *Encoder) encodeRz() bool {
	if e.stringAt(-2, "GARZ", "KURZ", "MARZ", "MERZ", "HERZ", "PERZ", "WARZ") ||
		e.stringAt(0, "RZANO", "RZOLA") || e.stringAt(-1, "ARZA", "ARZN") {
		return false
	}

	// 'yastrzemski' usually has 'z' silent in
	// united states, but should get 'X' in poland
	if e.stringAt(-4, "YASTRZEMSKI") {
		e.metaphAddAlt('R', 'X')
		e.idx++
		return true
	}

	// 'BRZEZINSKI' gets two pronunciations
	// in the united states, neither of which
	// are authentically polish
	if e.stringAt(-1, "BRZEZINSKI") {
		e.metaphAddStr("RS", "RJ")
		//skip of 2nd Z
		e.idx += 3
		return true
	}

	// 'z' in 'rz after voiceless consonant gets 'X'
	// in alternate polish style pronunciation
	if e.stringAt(-1, "TRZ", "PRZ", "KRZ") ||
		(e.stringAt(0, "RZ") && (e.isVowelAt(-1) || e.idx == 0)) {
		e.metaphAddStr("RS", "X")
		e.idx++
		return true
	}

	// 'z' in 'rz after voiceled consonant, vowel, or at
	// beginning gets 'J' in alternate polish style pronunciation
	if e.stringAt(-1, "BRZ", "DRZ", "GRZ") {
		e.metaphAddStr("RS", "J")
		e.idx++
		return true
	}

	return false
}

func (e *Encoder) testSilentR() bool {
	// test cases where 'R' is silent, either because the
	// word is from the french or because it is no longer pronounced.
	// e.g. "rogier", "monsieur", "surburban"
	if (e.idx == e.lastIdx &&
		e.stringAt(-2, "IER") &&
		// e.g. "metier"
		(e.stringAt(-5, "MET", "VIV", "LUC") ||
			// e.g. "cartier", "bustier"
			e.stringAt(-6, "CART", "DOSS", "FOUR", "OLIV", "BUST", "DAUM", "ATEL", "SONN",
				"CORM", "MERC", "PELT", "POIR", "BERN", "FORT", "GREN", "SAUC", "GAGN", "GAUT", "GRAN",
				"FORC", "MESS", "LUSS", "MEUN", "POTH", "HOLL", "CHEN") ||
			// e.g. "croupier"
			e.stringAt(-7, "CROUP", "TORCH", "CLOUT", "FOURN", "GAUTH", "TROTT", "DEROS", "CHART") ||
			// e.g. "chevalier"
			e.stringAt(-8, "CHEVAL", "LAVOIS", "PELLET", "SOMMEL", "TREPAN", "LETELL", "COLOMB") ||
			e.stringAt(-9, "CHARCUT") || e.stringAt(-10, "CHARPENT"))) ||
		e.stringAt(-2, "SURBURB", "WORSTED", "WORCESTER") ||
		e.stringAt(-7, "MONSIEUR") || e.stringAt(-6, "POITIERS") {

		return true
	}

	return false
}

//Encode '-re-" as 'AR' in contexts where this is the correct pronunciation
func (e *Encoder) encodeVowelReTransposition() bool {
	// -re inversion is just like
	// -le inversion
	// e.g. "fibre" => FABAR or "centre" => SANTAR
	if e.EncodeVowels && e.charNextIs('E') && len(e.in) > 3 &&
		!e.stringStart("OUTRE", "LIBRE", "ANDRE") && !e.stringExact("FRED", "TRES") &&
		!e.stringAt(-2, "LDRED", "LFRED", "NDRED", "NFRED", "NDRES", "TRES", "IFRED") &&
		!e.isVowelAt(-1) &&
		(e.idx+1 == e.lastIdx || e.stringAtEnd(2, "D", "S")) {

		e.metaphAddStr("AR", "AR")
		return true
	}

	return false
}

//650
func (e *Encoder) encodeS() {
	if e.encodeSkj() || e.encodeSpecialSw() || e.encodeSj() || e.encodeSilentFrenchSFinal() ||
		e.encodeSilentFrenchSInternal() || e.encodeIsl() || e.encodeStl() || e.encodeChristmas() ||
		e.encodeSthm() || e.encodeIsten() || e.encodeSugar() || e.encodeSh() || e.encodeSch() ||
		e.encodeSur() || e.encodeSu() || e.encodeSsio() || e.encodeSs() || e.encodeSia() ||
		e.encodeSio() || e.encodeAnglicisations() || e.encodeSc() || e.encodeSeiSuiSier() ||
		e.encodeSea() {
		return
	}

	e.metaphAdd('S')

	if e.stringAt(1, "S", "Z") && !e.stringAt(1, "SH") {
		e.idx++
	}
}

func (e *Encoder) encodeSkj() bool {
	if e.stringAt(0, "SKJO", "SKJU") && e.isVowelAt(3) {
		e.metaphAdd('X')
		e.idx += 2
		return true
	}
	return false
}

func (e *Encoder) encodeSpecialSw() bool {
	if e.idx == 0 {
		if e.namesBeginningWithSwThatGetAltSv() {
			e.metaphAddStr("S", "SV")
			e.idx++
			return true
		}

		if e.namesBeginningWithSwThatGetAlvXV() {
			e.metaphAddStr("S", "XV")
			e.idx++
			return true
		}
	}
	return false
}

func (e *Encoder) namesBeginningWithSwThatGetAltSv() bool {
	return e.stringStart("SWANSON", "SWENSON", "SWINSON", "SWENSEN", "SWOBODA",
		"SWIDERSKI", "SWARTHOUT", "SWEARENGIN")
}

func (e *Encoder) namesBeginningWithSwThatGetAlvXV() bool {
	return e.stringStart("SWART", "SWARTZ", "SWARTS", "SWIGER",
		"SWITZER", "SWANGER", "SWIGERT", "SWIGART", "SWIHART",
		"SWEITZER", "SWATZELL", "SWINDLER", "SWINEHART", "SWEARINGEN")
}

func (e *Encoder) encodeSj() bool {
	if e.stringStart("SJ") {
		e.metaphAdd('X')
		e.idx++
		return true
	}
	return false
}

func (e *Encoder) encodeSilentFrenchSFinal() bool {
	// "louis" is an exception because it gets two pronuncuations
	if e.stringExact("LOUIS") {
		e.metaphAddAlt('S', unicode.ReplacementChar)
		return true
	}

	if e.idx == e.lastIdx &&
		((e.stringStart("YVES", "ARKANSAS", "FRANCAIS", "CRUDITES", "BRUYERES",
			"DESCARTES", "DESCHUTES", "DESCHAMPS", "DESROCHES", "DESCHENES",
			"RENDEZVOUS", "CONTRETEMPS", "DESLAURIERS") ||
			e.stringExact("HORS") ||
			e.stringEnd("CAMUS", "YPRES",
				"MESNES", "DEBRIS", "BLANCS", "INGRES", "CANNES",
				"CHABLIS", "APROPOS", "JACQUES", "ELYSEES", "OEUVRES", "GEORGES", "DESPRES")) ||
			(e.stringAt(-2, "AI", "OI", "UI") && !e.stringStart("LOIS", "LUIS"))) {

		return true
	}
	return false
}

func (e *Encoder) encodeSilentFrenchSInternal() bool {
	// french words familiar to americans where internal s is silent
	return e.stringAt(-2, "MESNES", "DESCHAM", "DESPRES", "DESROCH", "DESROSI", "DESJARD", "DESMARA",
		"DESCHEN", "DESHOTE", "DESLAUR", "DESCARTES") ||
		e.stringAt(-5, "DUQUESNE", "DUCHESNE") ||
		e.stringAt(-3, "FRESNEL", "GROSVENOR") ||
		e.stringAt(-4, "LOUISVILLE") ||
		e.stringAt(-7, "BEAUCHESNE", "ILLINOISAN")
}

func (e *Encoder) encodeIsl() bool {
	// special cases 'island', 'isle', 'carlisle', 'carlysle'
	return (e.stringAt(-2, "LISL", "LYSL", "AISL") &&
		!e.stringAt(-3, "PAISLEY", "BAISLEY", "ALISLAM", "ALISLAH", "ALISLAA")) ||
		(e.idx == 1 && (e.stringAt(-1, "ISLE", "ISLAN") && !e.stringAt(-1, "ISLEY", "ISLER")))
}

func (e *Encoder) encodeStl() bool {
	// 'hustle', 'bustle', 'whistle'
	if (e.stringAt(0, "STLE", "STLI") && !e.stringAt(2, "LESS", "LIKE", "LINE")) ||
		e.stringAt(-3, "THISTLY", "BRISTLY", "GRISTLY") ||
		// e.g. "corpuscle"
		e.stringAt(-1, "USCLE") {

		// KRISTEN, KRYSTLE, CRYSTLE, KRISTLE all pronounce the 't'
		// also, exceptions where "-LING" is a nominalizing suffix
		if e.stringStart("KRISTEN", "KRYSTLE", "CRYSTLE", "KRISTLE", "CHRISTENSEN", "CHRISTENSON") ||
			e.stringAt(-3, "FIRSTLING") ||
			e.stringAt(-2, "NESTLING", "WESTLING") {
			e.metaphAddStr("ST", "ST")
			e.idx++
		} else {
			if e.EncodeVowels && e.charAt(3, 'E') && !e.charAt(4, 'R') &&
				!e.stringAt(3, "EY", "ETTE", "ETTA") {

				e.metaphAddStr("SAL", "SAL")
				e.flagAlInversion = true
			} else {
				e.metaphAddStr("SL", "SL")
			}
			e.idx += 2
		}
		return true
	}

	return false
}

func (e *Encoder) encodeChristmas() bool {
	if e.stringAt(-4, "CHRISTMA") {
		e.metaphAddStr("SM", "SM")
		e.idx += 2
		return true
	}
	return false
}

func (e *Encoder) encodeSthm() bool {
	// 'asthma', 'isthmus'
	if e.stringAt(0, "STHM") {
		e.metaphAddStr("SM", "SM")
		e.idx += 3
		return true
	}
	return false
}

func (e *Encoder) encodeIsten() bool {
	// 't' is silent in verb, pronounced in name
	if e.stringStart("CHRISTEN") {
		if rootOrInflections(e.in, "CHRISTEN") || e.stringStart("CHRISTENDOM") {
			e.metaphAddStr("S", "ST")
		} else {
			e.metaphAddStr("ST", "ST")
		}
		e.idx++
		return true
	}

	// e.g. 'glisten', 'listen'
	if e.stringAt(-2, "LISTEN", "RISTEN", "HASTEN", "FASTEN", "MUSTNT") || e.stringAt(-3, "MOISTEN") {
		e.metaphAdd('S')
		e.idx++
		return true
	}

	return false
}

func (e *Encoder) encodeSugar() bool {
	if e.stringAt(0, "SUGAR") {
		e.metaphAdd('X')
		return true
	}
	return false
}

func (e *Encoder) encodeSh() bool {
	if e.stringAt(0, "SH") {
		// exception
		if e.stringAt(-2, "CASHMERE") {
			e.metaphAdd('J')
			e.idx++
			return true
		}

		// combining forms, e.g. 'clotheshorse', 'woodshole'
		if e.idx > 0 &&
			(e.stringAtEnd(1, "HAP") ||
				// e.g. "hartsheim", "clothshorse"
				// e.g. "dishonor"
				e.stringAt(1, "HEIM", "HOEK", "HOLM", "HOLZ", "HOOD", "HEAD", "HEID",
					"HAAR", "HORS", "HOLE", "HUND", "HELM", "HAWK", "HILL", "HEART", "HATCH", "HOUSE", "HOUND", "HONOR") ||
				// e.g. "mishear"
				e.stringAtEnd(2, "EAR") ||
				// e.g. "hartshorn"
				(e.stringAt(2, "ORN") && !e.stringAt(-2, "UNSHORN")) ||
				// e.g. "newshour" but not "bashour", "manshour"
				(e.stringAt(1, "HOUR") && !e.stringStart("ASHOUR", "BASHOUR", "MANSHOUR")) ||
				// e.g. "dishonest", "grasshopper"
				e.stringAt(2, "ARMON", "ONEST", "ALLOW", "OLDER", "OPPER", "EIMER",
					"ANDLE", "ONOUR", "ABILLE", "UMANCE", "ABITUA")) {
			if !e.stringAt(-1, "S") {
				e.metaphAdd('S')
			}
		} else {
			e.metaphAdd('X')
		}

		e.idx++
		return true
	}

	return false
}

func (e *Encoder) encodeSch() bool {
	// these words were combining forms many centuries ago
	if e.stringAt(1, "CH") {
		if e.idx > 0 &&
			// e.g. "mischief", "escheat"
			(e.stringAt(3, "IEF", "EAT", "ANCE", "ARGE") ||
				e.stringStart("ESCHEW")) {

			e.metaphAdd('S')
			return true
		}

		// Schlesinger's rule
		// dutch, danish, italian, greek origin, e.g. "school", "schooner", "schiavone",
		// "schiz-"
		if (e.stringAt(3, "OO", "ER", "EN", "UY", "ED", "EM", "IA", "IZ", "IS", "OL") &&
			!e.stringAt(0, "SCHOLT", "SCHISL", "SCHERR")) ||
			e.stringAt(3, "ISZ") ||
			(e.stringAt(-1, "ESCHAT", "ASCHIN", "ASCHAL", "ISCHAE", "ISCHIA") &&
				!e.stringAt(-2, "FASCHING")) ||
			e.stringAtEnd(-1, "ESCHI") ||
			e.charAt(3, 'Y') {
			// e.g. "schermerhorn", "schenker", "schistose"

			if e.stringAt(3, "ER", "EN", "IS") &&
				(e.idx+4 == e.lastIdx || e.stringAt(3, "ENK", "ENB", "IST")) {

				e.metaphAddStr("X", "SK")
			} else {
				e.metaphAddStr("SK", "SK")
			}

			e.idx += 2
			return true
		} else {
			e.metaphAdd('X')
			e.idx += 2
			return true
		}
	}

	return false
}

func (e *Encoder) encodeSur() bool {
	// 'erasure', 'usury'
	if e.stringAt(1, "URE", "URA", "URY") {
		// 'sure', 'ensure'
		if e.idx == 0 || e.stringAt(-1, "N", "K") || e.stringAt(-2, "NO") {
			e.metaphAdd('X')
		} else {
			e.metaphAdd('J')
		}

		e.advanceCounter(1, 0)
		return true
	}
	return false
}

func (e *Encoder) encodeSu() bool {
	// 'sensuous', 'consensual'
	if e.stringAt(1, "UO", "UA") && e.idx != 0 {
		// exceptions e.g. "persuade"
		if e.stringAt(-1, "RSUA") {
			e.metaphAdd('S')
		} else if e.isVowelAt(-1) {
			// exceptions e.g. "casual"
			e.metaphAddAlt('J', 'S')
		} else {
			e.metaphAddAlt('X', 'S')
		}

		e.advanceCounter(2, 0)
		return true
	}
	return false
}

func (e *Encoder) encodeSsio() bool {
	if e.stringAt(1, "SION") {
		// "abcission"
		if e.stringAt(-2, "CI") {
			e.metaphAdd('J')
		} else if e.isVowelAt(-1) {
			// 'mission'
			e.metaphAdd('X')
		}

		e.advanceCounter(3, 1)
		return true
	}
	return false
}

func (e *Encoder) encodeSs() bool {
	// e.g. "russian", "pressure"
	// e.g. "hessian", "assurance"
	if e.stringAt(-1, "USSIA", "ESSUR", "ISSUR", "ISSUE", "ESSIAN", "ASSURE", "ASSURA", "ISSUAB", "ISSUAN", "ASSIUS") {
		e.metaphAdd('X')
		e.advanceCounter(2, 1)
		return true
	}
	return false
}

func (e *Encoder) encodeSia() bool {
	// e.g. "controversial", also "fuchsia", "ch" is silent
	if e.stringAt(-2, "CHSIA") || e.stringAt(-1, "RSIAL") {
		e.metaphAdd('X')
		e.advanceCounter(2, 0)
		return true
	}

	// names generally get 'X' where terms, e.g. "aphasia" get 'J'
	if (e.stringAtStart(-3, "ALESIA", "ALYSIA", "ALISIA", "STASIA") && !e.stringStart("ANASTASIA")) ||
		e.stringAt(-5, "THERESIA", "DIONYSIAN") {

		e.metaphAddAlt('X', 'S')
		e.advanceCounter(2, 0)
		return true
	}

	if e.stringAtEnd(0, "SIA", "SIAN") || e.stringAt(-5, "AMBROSIAL") {
		if (e.isVowelAt(-1) || e.stringAt(-1, "R")) &&
			// exclude compounds based on names, or french or greek words
			!(e.stringStart("JAMES", "NICOS", "PEGAS", "PEPYS",
				"HOBBES", "HOLMES", "JAQUES", "KEYNES",
				"MALTHUS", "HOMOOUS", "MAGLEMOS", "HOMOIOUS",
				"LEVALLOIS", "TARDENOIS") || e.stringAt(-4, "ALGES")) {

			e.metaphAdd('J')
		} else {
			e.metaphAdd('S')
		}

		e.advanceCounter(1, 0)
		return true
	}

	return false
}

func (e *Encoder) encodeSio() bool {
	// special case, irish name
	if e.stringStart("SIOBHAN") {
		e.metaphAdd('X')
		e.advanceCounter(2, 0)
		return true
	}
	if e.stringAt(1, "ION") {
		// e.g. "vision", "version"
		if e.isVowelAt(-1) || e.stringAt(-2, "ER", "UR") {
			e.metaphAdd('J')
		} else {
			// e.g. "declension"
			e.metaphAdd('X')
		}
		e.advanceCounter(2, 0)
		return true
	}
	return false
}

func (e *Encoder) encodeAnglicisations() bool {
	// german & anglicisations, e.g. 'smith' match 'schmidt', 'snider' match
	// 'schneider'
	// also, -sz- in slavic language altho in hungarian it is pronounced 's'

	if e.stringAtStart(0, "SM", "SN", "SL") || e.stringAt(1, "Z") {
		e.metaphAddAlt('S', 'X')

		// eat redundant 'Z'
		if e.stringAt(1, "Z") {
			e.idx++
		}

		return true
	}

	return false
}

func (e *Encoder) encodeSc() bool {
	if e.stringAt(0, "SC") {
		// exception 'viscount'
		if e.stringAt(-2, "VISCOUNT") {
			return true
		}

		// encode "-SC<front vowel>-"
		if e.stringAt(2, "I", "E", "Y") {
			// e.g. "conscious"
			// e.g. "prosciutto"
			if e.stringAt(2, "IUT", "IOUS") || e.stringAt(-2, "FASCIS") ||
				e.stringAt(-3, "CONSCIEN", "CRESCEND", "CONSCION") || e.stringAt(-4, "OMNISCIEN") {
				e.metaphAdd('X')
			} else if e.stringAt(0, "SCIVV", "SCIRO", "SCIPIO", "SCEPTIC", "SCEPSIS") ||
				e.stringAt(-2, "PISCITELLI") {
				e.metaphAddStr("SK", "SK")
			} else {
				e.metaphAdd('S')
			}

			e.idx++
			return true
		}

		e.metaphAddStr("SK", "SK")
		e.idx++
		return true
	}
	return false
}

func (e *Encoder) encodeSeiSuiSier() bool {
	// "nausea" by itself has => NJ as a more likely encoding. Other forms
	// using "nause-" (see Encode_SEA()) have X or S as more familiar
	// pronounciations
	if e.stringAtEnd(-3, "NAUSEA") ||
		e.stringAt(-2, "CASUI") ||
		(e.stringAt(-1, "OSIER", "ASIER") &&
			!(e.stringStart("OSIER", "EASIER") || e.stringAt(-2, "ROSIER", "MOSIER"))) {

		e.metaphAddAlt('J', 'X')
		e.advanceCounter(2, 0)
		return true
	}

	return false
}

func (e *Encoder) encodeSea() bool {
	//TODO: bug?  NAUSEO and not NAUSEAT?
	if e.stringExact("SEAN") || (e.stringAt(-3, "NAUSEO") && !e.stringAt(-3, "NAUSEAT")) {
		e.metaphAdd('X')
		e.advanceCounter(2, 0)
		return true
	}
	return false
}

func (e *Encoder) encodeT() {
	if e.encodeTInitial() || e.encodeTch() || e.encodeSilentFrenchT() ||
		e.encodeTunTulTuaTuo() || e.encodeTueTeuTeouTulTie() || e.encodeTurTiuSuffixes() ||
		e.encodeTi() || e.encodeTient() || e.encodeTsch() || e.encodeTzsch() ||
		e.encodeThPronouncedSeparately() || e.encodeTth() || e.encodeTh() {
		return
	}

	if e.stringAt(1, "T", "D") {
		e.idx++
	}
	e.metaphAdd('T')
}

func (e *Encoder) encodeTInitial() bool {
	if e.idx == 0 {
		// americans usually pronounce "tzar" as "zar"
		if e.stringAt(1, "SAR", "ZAR") {
			return true
		}

		// old 'École française d'Extrême-Orient' chinese pinyin where 'ts-' => 'X'
		if e.stringExact("TSO", "TSA", "TSU", "TSAO", "TSAI", "TSING", "TSANG") {
			e.metaphAdd('X')
			e.advanceCounter(2, 1)
			return true
		}

		// "TS<vowel>-" at start can be pronounced both with and without 'T'
		if e.charNextIs('S') && e.isVowelAt(2) {
			e.metaphAddStr("TS", "S")
			e.advanceCounter(2, 1)
			return true
		}

		// e.g. "Tjaarda"
		if e.charNextIs('J') {
			e.metaphAdd('X')
			e.advanceCounter(2, 1)
			return true
		}

		if e.stringExact("THU") || e.stringAt(1, "HAI", "HUY", "HAO", "HYME", "HYMY", "HANH", "HERES") {
			e.metaphAdd('T')
			e.advanceCounter(2, 1)
			return true
		}
	}

	return false
}

func (e *Encoder) encodeTch() bool {
	if e.stringAt(1, "CH") {
		e.metaphAdd('X')
		e.idx += 2
		return true
	}
	return false
}

func (e *Encoder) encodeSilentFrenchT() bool {
	// french silent T familiar to americans
	return (e.stringAtEnd(-4, "MONET", "GENET", "CHAUT") ||
		e.stringAt(-2, "POTPOURRI") ||
		e.stringAt(-3, "MORTGAGE", "BOATSWAIN") ||
		e.stringAt(-4, "BERET", "BIDET", "FILET", "DEBUT", "DEPOT", "PINOT", "TAROT") ||
		e.stringAt(-5, "BALLET", "BUFFET", "CACHET", "CHALET", "ESPRIT", "RAGOUT", "GOULET", "CHABOT", "BENOIT") ||
		e.stringAt(-6, "GOURMET", "BOUQUET", "CROCHET", "CROQUET", "PARFAIT", "PINCHOT", "CABARET", "PARQUET", "RAPPORT", "TOUCHET", "COURBET", "DIDEROT") ||
		e.stringAt(-7, "ENTREPOT", "CABERNET", "DUBONNET", "MASSENET", "MUSCADET", "RICOCHET", "ESCARGOT") ||
		e.stringAt(-8, "SOBRIQUET", "CABRIOLET", "CASSOULET", "OUBRIQUET", "CAMEMBERT")) &&
		!e.stringAt(1, "AN", "RY", "IC", "OM", "IN")
}

func (e *Encoder) encodeTunTulTuaTuo() bool {
	// e.g. "fortune", "fortunate"
	if e.stringAt(-3, "FORTUN") ||
		// e.g. "capitulate"
		(e.stringAt(0, "TUL") && e.isVowelAt(-1) && e.isVowelAt(3)) ||
		// e.g. "obituary", "barbituate"
		e.stringAt(-2, "BITUA", "BITUE") ||
		// e.g. "actual"
		(e.idx > 1 && e.stringAt(0, "TUA", "TUO")) {

		e.metaphAddAlt('X', 'T')
		return true
	}
	return false
}

func (e *Encoder) encodeTueTeuTeouTulTie() bool {
	if e.stringAt(1, "UENT") ||
		e.stringAt(-4, "RIGHTEOUS") ||
		e.stringAt(-3, "STATUTE", "AMATEUR", "STATUTOR") ||
		// e.g. "blastula", "pasteur"
		e.stringAt(-1, "NTULE", "NTULA", "STULE", "STULA", "STEUR") ||
		// e.g. "statue"
		e.stringAtEnd(0, "TUE") ||
		// e.g. "constituency"
		e.stringAt(0, "TUENC") ||
		// e.g. "patience"
		e.stringAtEnd(0, "TIENCE") {

		e.metaphAddAlt('X', 'T')
		e.advanceCounter(1, 0)
		return true
	}

	return false
}

func (e *Encoder) encodeTurTiuSuffixes() bool {
	// 'adventure', 'musculature'
	if e.idx > 0 && e.stringAt(1, "URE", "URA", "URI", "URY", "URO", "IUS") {
		// exceptions e.g. 'tessitura', mostly from romance languages
		if (e.stringAtEnd(1, "URA", "URO") && !e.stringAt(-3, "VENTURA")) ||
			// e.g. "kachaturian", "hematuria"
			e.stringAt(1, "URIA") {

			e.metaphAdd('T')
		} else {
			e.metaphAddAlt('X', 'T')
		}

		e.advanceCounter(1, 0)
		return true
	}
	return false
}

func (e *Encoder) encodeTi() bool {
	// '-tio-', '-tia-', '-tiu-'
	// except combining forms where T already pronounced e.g 'rooseveltian'
	if (e.stringAt(1, "IO") && !e.stringAt(-1, "ETIOL")) ||
		e.stringAt(1, "IAL") ||
		e.stringAt(-1, "RTIUM", "ATIUM") ||
		((e.stringAt(1, "IAN") && e.idx > 0) &&
			!(e.stringAt(-4, "FAUSTIAN") || e.stringAt(-5, "PROUSTIAN") ||
				e.stringAt(-2, "TATIANA") || e.stringAt(-3, "KANTIAN", "GENTIAN") ||
				e.stringAt(-8, "ROOSEVELTIAN")) ||
			(e.stringAtEnd(0, "TIA") &&
				// exceptions to above rules where the pronounciation is usually X
				!(e.stringAt(-3, "HESTIA", "MASTIA") ||
					e.stringAt(-2, "OSTIA") || e.stringStart("TIA") ||
					e.stringAt(-5, "IZVESTIA"))) ||
			e.stringAt(1, "IATE", "IATI", "IABL", "IATO", "IARY") ||
			e.stringAt(-5, "CHRISTIAN")) {

		if e.stringAtStart(-2, "ANTI") || e.stringStart("PATIO", "PITIA", "DUTIA") {
			e.metaphAdd('T')
		} else if e.stringAt(-4, "EQUATION") {
			e.metaphAdd('J')
		} else if e.stringAt(0, "TION") {
			e.metaphAdd('X')
		} else if e.stringStart("KATIA", "LATIA") {
			e.metaphAddAlt('T', 'X')
		} else {
			e.metaphAddAlt('X', 'T')
		}

		e.advanceCounter(2, 0)
		return true
	}

	return false
}

func (e *Encoder) encodeTient() bool {
	// e.g. 'patient'
	if e.stringAt(1, "IENT") {
		e.metaphAddAlt('X', 'T')
		e.advanceCounter(2, 0)
		return true
	}
	return false
}

func (e *Encoder) encodeTsch() bool {
	// 'deutsch'
	if e.stringAt(0, "TSCH") &&
		// combining forms in german where the 'T' is pronounced seperately
		!e.stringAt(-3, "WELT", "KLAT", "FEST") {

		// pronounced the same as "ch" in "chit" => X
		e.metaphAdd('X')
		e.idx += 3
		return true
	}
	return false
}

func (e *Encoder) encodeTzsch() bool {
	// 'neitzsche'
	if e.stringAt(0, "TZSCH") {
		e.metaphAdd('X')
		e.idx += 4
		return true
	}
	return false
}

func (e *Encoder) encodeThPronouncedSeparately() bool {
	// 'adulthood', 'bithead', 'apartheid'
	if (e.idx > 0 && e.stringAt(1, "HOOD", "HEAD", "HEID", "HAND", "HILL", "HOLD", "HAWK", "HEAP", "HERD",
		"HOLE", "HOOK", "HUNT", "HUMO", "HAUS", "HOFF", "HARD") && !e.stringAt(-3, "SOUTH", "NORTH")) ||
		e.stringAt(1, "HOUSE", "HEART", "HASTE", "HYPNO", "HEQUE") ||
		// watch out for greek root "-thallic"
		(e.stringAtEnd(1, "HALL") && !e.stringAt(-3, "SOUTH", "NORTH")) ||
		(e.stringAtEnd(1, "HAM") && !e.stringStart("GOTHAM", "WITHAM", "LATHAM", "BENTHAM", "WALTHAM", "WORTHAM", "GRANTHAM")) ||
		(e.stringAt(1, "HATCH") && !(e.idx == 0 || e.stringAt(-2, "UNTHATCH"))) ||
		e.stringAt(-3, "GOETHE", "WARTHOG") ||
		// and some special cases where "-TH-" is usually pronounced 'T'
		e.stringAt(-2, "ESTHER", "NATHALIE") {

		//special case
		if e.stringAt(-3, "POSTHUM") {
			e.metaphAdd('X')
		} else {
			e.metaphAdd('T')
		}
		e.idx++
		return true
	}

	return false
}

func (e *Encoder) encodeTth() bool {
	// 'matthew' vs. 'outthink'
	if e.stringAt(0, "TTH") {
		if e.stringAt(-2, "MATTH") {
			e.metaphAdd('0')
		} else {
			e.metaphAddStr("T0", "T0")
		}
		e.idx += 2
		return true
	}

	return false
}

func (e *Encoder) encodeTh() bool {
	if e.stringAt(0, "TH") {
		// '-clothes-'
		if e.stringAt(-3, "CLOTHES") {
			// vowel already encoded so skip right to S
			e.idx += 2
			return true
		}

		// special case "thomas", "thames", "beethoven" or germanic words
		if e.stringAt(2, "OMAS", "OMPS", "OMPK", "OMSO", "OMSE", "AMES", "OVEN", "OFEN", "ILDA", "ILDE") ||
			e.stringExact("THOM", "THOMS") ||
			e.stringStart("SCH", "VAN ", "VON ") {

			e.metaphAdd('T')
		} else {
			// give an 'etymological' 2nd
			// encoding for "smith"
			if e.stringStart("SM") {
				e.metaphAddAlt('0', 'T')
			} else {
				e.metaphAdd('0')
			}
		}

		e.idx++
		return true
	}
	return false
}

func (e *Encoder) encodeV() {
	if e.charNextIs('V') {
		e.idx++
	}
	e.metaphAddExactApprox("V", "F")
}

func (e *Encoder) encodeW() {
	if e.encodeSilentWAtBeginning() || e.encodeWitzWicz() || e.encodeWr() ||
		e.encodeInitialWVowel() || e.encodeWh() || e.encodeEasternEuropeanW() {
		return
	}

	// e.g. 'zimbabwe'
	if e.EncodeVowels && e.stringAtEnd(0, "WE") {
		e.metaphAdd('A')
	}
}

func (e *Encoder) encodeSilentWAtBeginning() bool {
	return e.stringAtStart(0, "WR")
}

func (e *Encoder) encodeWitzWicz() bool {
	// polish e.g. 'filipowicz'
	if e.stringAtEnd(0, "WICZ", "WITZ") {
		if e.EncodeVowels {
			// don't dupe A's
			if len(e.primBuf) > 0 && e.primBuf[len(e.primBuf)-1] == 'A' {
				e.metaphAddStr("TS", "FAX")
			} else {
				e.metaphAddStr("ATS", "FAX")
			}
		} else {
			e.metaphAddStr("TS", "FX")
		}

		e.idx += 3
		return true
	}
	return false
}

func (e *Encoder) encodeWr() bool {
	// can also be in middle of word
	if e.stringAt(0, "WR") {
		e.metaphAdd('R')
		e.idx++
		return true
	}
	return false
}

func (e *Encoder) encodeInitialWVowel() bool {
	if e.idx == 0 && e.isVowelAt(1) {
		// Witter should match Vitter
		if e.germanicOrSlavicNameBeginningWithW() {
			if e.EncodeVowels {
				e.metaphAddExactApproxAlt("A", "VA", "A", "FA")
			} else {
				e.metaphAddExactApproxAlt("A", "V", "A", "F")
			}
		} else {
			e.metaphAdd('A')
		}

		e.idx = e.skipVowels(e.idx + 1)
		return true
	}

	return false
}

func (e *Encoder) encodeWh() bool {
	if e.stringAt(0, "WH") {
		// cases where it is pronounced as H
		// e.g. 'who', 'whole'
		if e.charAt(2, 'O') && !e.stringAt(2, "OA", "OP", "OOP", "OMP", "ORL", "ORT", "OOSH") {
			e.metaphAdd('H')
			e.advanceCounter(2, 1)
			return true
		}

		// combining forms, e.g. 'hollowhearted', 'rawhide'
		if e.stringAt(2, "IDE", "ARD", "EAD", "AWK", "ERD", "OOK", "AND", "OLE", "OOD",
			"EART", "OUSE", "OUND", "AMMER") {
			e.metaphAdd('H')
			e.idx++
			return true
		}

		if e.idx == 0 {
			e.metaphAdd('A')
			e.idx = e.skipVowels(e.idx + 2)
			return true
		}

		e.idx++
		return true
	}

	return false
}

func (e *Encoder) encodeEasternEuropeanW() bool {
	// Arnow should match Arnoff
	if (e.idx == e.lastIdx && e.isVowelAt(-1)) ||
		e.stringAt(-1, "EWSKI", "EWSKY", "OWSKI", "OWSKY") ||
		e.stringAtEnd(0, "WIAK", "WICKI", "WACKI") ||
		e.stringStart("SCH") {

		e.metaphAddExactApproxAlt("", "V", "", "F")
		return true
	}
	return false
}

func (e *Encoder) germanicOrSlavicNameBeginningWithW() bool {
	return e.stringStart("WEE", "WIX", "WAX",
		"WOLF", "WEIS", "WAHL", "WALZ", "WEIL", "WERT", "WINE", "WILK", "WALT", "WOLL",
		"WADA", "WULF", "WEHR", "WURM", "WYSE", "WENZ", "WIRT", "WOLK", "WEIN", "WYSS", "WASS", "WANN",
		"WINT", "WINK", "WILE", "WIKE", "WIER", "WELK", "WISE",
		"WIRTH", "WIESE", "WITTE", "WENTZ", "WOLFF", "WENDT", "WERTZ", "WILKE", "WALTZ",
		"WEISE", "WOOLF", "WERTH", "WEESE", "WURTH", "WINES", "WARGO", "WIMER", "WISER", "WAGER",
		"WILLE", "WILDS", "WAGAR", "WERTS", "WITTY", "WIENS", "WIEBE", "WIRTZ", "WYMER", "WULFF",
		"WIBLE", "WINER", "WIEST", "WALKO", "WALLA", "WEBRE", "WEYER", "WYBLE", "WOMAC", "WILTZ",
		"WURST", "WOLAK", "WELKE", "WEDEL", "WEIST", "WYGAN", "WUEST", "WEISZ", "WALCK", "WEITZ",
		"WYDRA", "WANDA", "WILMA", "WEBER",
		"WETZEL", "WEINER", "WENZEL", "WESTER", "WALLEN", "WENGER", "WALLIN", "WEILER",
		"WIMMER", "WEIMER", "WYRICK", "WEGNER", "WINNER", "WESSEL", "WILKIE", "WEIGEL", "WOJCIK",
		"WENDEL", "WITTER", "WIENER", "WEISER", "WEXLER", "WACKER", "WISNER", "WITMER", "WINKLE",
		"WELTER", "WIDMER", "WITTEN", "WINDLE", "WASHER", "WOLTER", "WILKEY", "WIDNER", "WARMAN",
		"WEYANT", "WEIBEL", "WANNER", "WILKEN", "WILTSE", "WARNKE", "WALSER", "WEIKEL", "WESNER",
		"WITZEL", "WROBEL", "WAGNON", "WINANS", "WENNER", "WOLKEN", "WILNER", "WYSONG", "WYCOFF",
		"WUNDER", "WINKEL", "WIDMAN", "WELSCH", "WEHNER", "WEIGLE", "WETTER", "WUNSCH", "WHITTY",
		"WAXMAN", "WILKER", "WILHAM", "WITTIG", "WITMAN", "WESTRA", "WEHRLE", "WASSER", "WILLER",
		"WEGMAN", "WARFEL", "WYNTER", "WERNER", "WAGNER", "WISSER",
		"WISEMAN", "WINKLER", "WILHELM", "WELLMAN", "WAMPLER", "WACHTER", "WALTHER",
		"WYCKOFF", "WEIDNER", "WOZNIAK", "WEILAND", "WILFONG", "WIEGAND", "WILCHER", "WIELAND",
		"WILDMAN", "WALDMAN", "WORTMAN", "WYSOCKI", "WEIDMAN", "WITTMAN", "WIDENER", "WOLFSON",
		"WENDELL", "WEITZEL", "WILLMAN", "WALDRUP", "WALTMAN", "WALCZAK", "WEIGAND", "WESSELS",
		"WIDEMAN", "WOLTERS", "WIREMAN", "WILHOIT", "WEGENER", "WOTRING", "WINGERT", "WIESNER",
		"WAYMIRE", "WHETZEL", "WENTZEL", "WINEGAR", "WESTMAN", "WYNKOOP", "WALLICK", "WURSTER",
		"WINBUSH", "WILBERT", "WALLACH", "WYNKOOP", "WALLICK", "WURSTER", "WINBUSH", "WILBERT",
		"WALLACH", "WEISSER", "WEISNER", "WINDERS", "WILLMON", "WILLEMS", "WIERSMA", "WACHTEL",
		"WARNICK", "WEIDLER", "WALTRIP", "WHETSEL", "WHELESS", "WELCHER", "WALBORN", "WILLSEY",
		"WEINMAN", "WAGAMAN", "WOMMACK", "WINGLER", "WINKLES", "WIEDMAN", "WHITNER", "WOLFRAM",
		"WARLICK", "WEEDMAN", "WHISMAN", "WINLAND", "WEESNER", "WARTHEN", "WETZLER", "WENDLER",
		"WALLNER", "WOLBERT", "WITTMER", "WISHART", "WILLIAM",
		"WESTPHAL", "WICKLUND", "WEISSMAN", "WESTLUND", "WOLFGANG", "WILLHITE", "WEISBERG",
		"WALRAVEN", "WOLFGRAM", "WILHOITE", "WECHSLER", "WENDLING", "WESTBERG", "WENDLAND", "WININGER",
		"WHISNANT", "WESTRICK", "WESTLING", "WESTBURY", "WEITZMAN", "WEHMEYER", "WEINMANN", "WISNESKI",
		"WHELCHEL", "WEISHAAR", "WAGGENER", "WALDROUP", "WESTHOFF", "WIEDEMAN", "WASINGER", "WINBORNE",
		"WHISENANT", "WEINSTEIN", "WESTERMAN", "WASSERMAN", "WITKOWSKI", "WEINTRAUB",
		"WINKELMAN", "WINKFIELD", "WANAMAKER", "WIECZOREK", "WIECHMANN", "WOJTOWICZ", "WALKOWIAK",
		"WEINSTOCK", "WILLEFORD", "WARKENTIN", "WEISINGER", "WINKLEMAN", "WILHEMINA",
		"WISNIEWSKI", "WUNDERLICH", "WHISENHUNT", "WEINBERGER", "WROBLEWSKI", "WAGUESPACK",
		"WEISGERBER", "WESTERVELT", "WESTERLUND", "WASILEWSKI", "WILDERMUTH", "WESTENDORF",
		"WESOLOWSKI", "WEINGARTEN", "WINEBARGER", "WESTERBERG", "WANNAMAKER", "WEISSINGER",
		"WALDSCHMIDT", "WEINGARTNER", "WINEBRENNER",
		"WOLFENBARGER", "WOJCIECHOWSKI")
}

func (e *Encoder) encodeX() {
	if e.encodeInitialX() || e.encodeGreekX() || e.encodeXSpecialCases() ||
		e.encodeXToH() || e.encodeXVowel() || e.encodeFrenchXFinal() {
		return
	}

	// eat redundant 'X' or other redundant cases
	// e.g. "excite", "exceed"
	if e.stringAt(1, "X", "Z", "S", "CE", "CE") {
		e.idx++
	}
}

func (e *Encoder) encodeInitialX() bool {
	// current chinese pinyin spelling
	if e.stringStart("XU", "XIA", "XIO", "XIE") {
		e.metaphAdd('X')
		return true
	}

	if e.idx == 0 {
		e.metaphAdd('S')
		return true
	}

	return false
}

// 'xylophone', xylem', 'xanthoma', 'xeno-'
func (e *Encoder) encodeGreekX() bool {
	if e.stringAt(1, "YLO", "YLE", "ENO", "ANTH") {
		e.metaphAdd('S')
		return true
	}
	return false
}

//Encode special cases, "LUXUR-", "Texeira"
func (e *Encoder) encodeXSpecialCases() bool {
	if e.stringAt(-2, "LUXUR") {
		e.metaphAddExactApprox("GJ", "KJ")
		return true
	}

	if e.stringStart("TEXEIRA", "TEIXEIRA") {
		e.metaphAdd('X')
		return true
	}
	return false
}

//Encode special case where americans know the proper mexican indian
//pronounciation of this name
func (e *Encoder) encodeXToH() bool {
	if e.stringAt(-2, "OAXACA") || e.stringAt(-3, "QUIXOTE") {
		e.metaphAdd('H')
		return true
	}
	return false
}

func (e *Encoder) encodeXVowel() bool {
	// e.g. "sexual", "connexion" (british), "noxious"
	if e.stringAt(1, "UAL", "ION", "IOU") {
		e.metaphAddStr("KX", "KS")
		e.advanceCounter(2, 0)
		return true
	}
	return false
}

func (e *Encoder) encodeFrenchXFinal() bool {
	if !(e.idx == e.lastIdx && (e.stringAt(-3, "IAU", "EAU", "IEU") ||
		e.stringAt(-2, "AI", "AU", "OU", "OI", "EU"))) {

		e.metaphAddStr("KS", "KS")
		//no return true?
	}
	return false
}

func (e *Encoder) encodeZ() {
	if e.encodeZz() || e.encodeZuZierZs() || e.encodeFrenchEz() || e.encodeGermanZ() || e.encodeZh() {
		return
	}

	e.metaphAdd('S')

	// eat redundant 'Z'
	if e.charNextIs('Z') {
		e.idx++
	}
}

//Encode cases of "-ZZ-" where it is obviously part of an italian word where
//"-ZZ-" is pronounced as TS
func (e *Encoder) encodeZz() bool {
	// "abruzzi", 'pizza'
	if e.charNextIs('Z') &&
		(e.stringAtEnd(2, "I", "O", "A") || e.stringAt(-2, "MOZZARELL", "PIZZICATO", "PUZZONLAN")) {

		e.metaphAddStr("TS", "S")
		e.idx++
		return true
	}
	return false
}

func (e *Encoder) encodeZuZierZs() bool {
	if (e.idx == 1 && e.stringAt(-1, "AZUR")) ||
		(e.stringAt(0, "ZIER") && !e.stringAt(-2, "VIZIER")) ||
		e.stringAt(0, "ZSA") {

		e.metaphAddAlt('J', 'S')

		if e.stringAt(0, "ZSA") {
			e.idx++
		}
		return true
	}
	return false
}

//Encode cases where americans recognize "-EZ" as part of a french word where Z
//not pronounced
func (e *Encoder) encodeFrenchEz() bool {
	if (e.idx == 3 && e.stringAt(-3, "CHEZ")) ||
		e.stringAt(-5, "RENDEZ") {
		return true
	}

	return false
}

//Encode cases where "-Z-" is in a german word where Z => TS in german
func (e *Encoder) encodeGermanZ() bool {
	if e.stringExact("NAZI") ||
		e.stringAt(-2, "NAZIFY", "MOZART") ||
		e.stringAt(-3, "HOLZ", "HERZ", "MERZ", "FITZ", "HERZOG") ||
		(e.stringAt(-3, "GANZ") && !e.isVowelAt(1)) ||
		e.stringAt(-4, "STOLZ", "PRINZ", "VENEZIA") ||
		// german words containing with "sch" but not schlimazel, schmooze
		(e.stringContains("SCH") && !e.stringEnd("IZE", "OZE", "ZEL")) ||
		(e.idx > 0 && e.stringAt(0, "ZEIT")) ||
		e.stringAt(-3, "WEIZ") {

		if e.idx > 0 && e.charAt(-1, 'T') {
			e.metaphAdd('S')
		} else {
			e.metaphAddStr("TS", "TS")
		}
		return true
	}

	return false
}

func (e *Encoder) encodeZh() bool {
	// chinese pinyin e.g. 'zhao', also english "phonetic spelling"
	if e.charNextIs('H') {
		e.metaphAdd('J')
		e.idx++
		return true
	}
	return false
}

func (e *Encoder) encodeVowels() {

	if e.idx == 0 {
		// all init vowels map to 'A'
		// as of Double Metaphone
		e.metaphAdd('A')
	} else if e.EncodeVowels {
		if !e.charAt(0, 'E') {
			if e.encodeSkipSilentUe() {
				return
			}
			if e.encodeOSilent() {
				return
			}
			// encode all vowels and
			// diphthongs to the same value
			e.metaphAdd('A')

		} else {
			e.encodeEPronounced()
		}
	}

	if !(!e.isVowelAt(-2) && e.stringAt(-1, "LEWA", "LEWO", "LEWI")) {
		e.idx = e.skipVowels(e.idx + 1)
	}
}

func (e *Encoder) encodeSkipSilentUe() bool {
	// always silent except for cases listed below
	if (e.stringAt(-1, "QUE", "GUE") &&
		!e.stringStart("RISQUE", "PIROGUE", "ENRIQUE", "BARBEQUE", "PALENQUE", "APPLIQUE", "COMMUNIQUE") &&
		!e.stringAt(-3, "ARGUE", "SEGUE")) &&
		e.idx > 1 &&
		((e.idx+1 == e.lastIdx) || e.stringStart("JACQUES")) {

		e.idx = e.skipVowels(e.idx)
		return true
	}
	return false
}

// Encodes cases where non-initial 'e' is pronounced, taking
// care to detect unusual cases from the greek.
// Only executed if non initial vowel encoding is turned on
func (e *Encoder) encodeEPronounced() {
	// special cases with two pronunciations
	// 'agape' 'lame' 'resume'
	if e.stringExact("LAME", "SAKE", "PATE", "AGAPE") ||
		(e.stringStart("RESUME") && e.idx == 5) {

		e.metaphAddAlt(unicode.ReplacementChar, 'A')
		return
	}

	// special case "inge" => 'INGA', 'INJ'
	if e.stringExact("INGE") {
		e.metaphAddAlt('A', unicode.ReplacementChar)
		return
	}

	// special cases with two pronunciations
	// special handling due to the difference in
	// the pronunciation of the '-D'
	if e.idx == 5 && e.stringStart("BLESSED", "LEARNED") {
		e.metaphAddExactApproxAlt("D", "AD", "T", "AT")
		e.idx++
		return
	}

	// encode all vowels and diphthongs to the same value
	if (!e.encodeESilent() && !e.flagAlInversion && !e.encodeSilentInternalE()) ||
		e.encodeEPronouncedExceptions() {

		e.metaphAdd('A')
	}

	// now that we've visited the vowel in question
	e.flagAlInversion = false
}

func (e *Encoder) encodeOSilent() bool {
	// if "iron" at beginning or end of word and not "irony"
	if e.charAt(0, 'O') && e.stringAt(-2, "IRON") {
		if (e.stringStart("IRON") || e.stringAtEnd(-2, "IRON")) && !e.stringAt(-2, "IRONIC") {
			return true
		}
	}

	return false
}

func (e *Encoder) encodeESilent() bool {
	if e.encodeEPronouncedAtEnd() {
		return false
	}

	// 'e' silent when last letter, altho
	if e.idx == e.lastIdx ||
		// also silent if before plural 's'
		// or past tense or participle 'd', e.g.
		// 'grapes' and 'banished' => PNXT
		(e.idx > 1 && e.idx+1 == e.lastIdx && e.stringAt(1, "S", "D") &&
			// and not e.g. "nested", "rises", or "pieces" => RASAS
			!(e.stringAt(-1, "TED", "SES", "CES") ||
				e.stringStart("ABED", "IMED", "JARED", "AHMED", "HAMED", "JAVED",
					"NORRED", "MEDVED", "MERCED", "ALLRED", "KHALED", "RASHED", "MASJED",
					"MOHAMED", "MOHAMMED", "MUHAMMED", "MOUHAMED", "ANTIPODES", "ANOPHELES"))) ||
		// e.g.  'wholeness', 'boneless', 'barely'
		e.stringAtEnd(1, "NESS", "LESS") ||
		(e.stringAtEnd(1, "LY") && !e.stringStart("CICELY")) {

		return true
	}
	return false
}

// Tests for words where an 'E' at the end of the word
// is pronounced
//
// special cases, mostly from the greek, spanish, japanese,
// italian, and french words normally having an acute accent.
// also, pronouns and articles
//
// Many Thanks to ali, QuentinCompson, JeffCO, ToonScribe, Xan,
// Trafalz, and VictorLaszlo, all of them atriots from the Eschaton,
// for all their fine contributions!
func (e *Encoder) encodeEPronouncedAtEnd() bool {
	if e.idx == e.lastIdx &&
		(e.stringAt(-6, "STROPHE") ||
			// if a vowel is before the 'E', vowel eater will have eaten it.
			//otherwise, consonant + 'E' will need 'E' pronounced
			len(e.in) == 2 ||
			(len(e.in) == 3 && !e.isVowelAt(-e.idx)) ||
			// these german name endings can be relied on to have the 'e' pronounced
			(e.stringAtEnd(-2, "BKE", "DKE", "FKE", "KKE", "LKE", "NKE", "MKE", "PKE", "TKE", "VKE", "ZKE") &&
				!e.stringStart("FINKE", "FUNKE", "FRANKE")) ||
			e.stringAtEnd(-4, "SCHKE") ||
			e.stringExact("ACME", "NIKE", "CAFE", "RENE", "LUPE", "JOSE", "ESME",
				"LETHE", "CADRE", "TILDE", "SIGNE", "POSSE", "LATTE", "ANIME", "DOLCE", "CROCE",
				"ADOBE", "OUTRE", "JESSE", "JAIME", "JAFFE", "BENGE", "RUNGE",
				"CHILE", "DESME", "CONDE", "URIBE", "LIBRE", "ANDRE",
				"HECATE", "PSYCHE", "DAPHNE", "PENSKE", "CLICHE", "RECIPE",
				"TAMALE", "SESAME", "SIMILE", "FINALE", "KARATE", "RENATE", "SHANTE",
				"OBERLE", "COYOTE", "KRESGE", "STONGE", "STANGE", "SWAYZE", "FUENTE",
				"SALOME", "URRIBE",
				"ECHIDNE", "ARIADNE", "MEINEKE", "PORSCHE", "ANEMONE", "EPITOME",
				"SYNCOPE", "SOUFFLE", "ATTACHE", "MACHETE", "KARAOKE", "BUKKAKE",
				"VICENTE", "ELLERBE", "VERSACE",
				"PENELOPE", "CALLIOPE", "CHIPOTLE", "ANTIGONE", "KAMIKAZE", "EURIDICE",
				"YOSEMITE", "FERRANTE",
				"HYPERBOLE", "GUACAMOLE", "XANTHIPPE",
				"SYNECDOCHE")) {

		return true
	}

	return false
}

func (e *Encoder) encodeSilentInternalE() bool {
	// 'olesen' but not 'olen'	RAKE BLAKE
	if (e.stringStart("OLE") && e.encodeESuffix(3)) ||
		(e.stringStart("BARE", "FIRE", "FORE", "GATE", "HAGE", "HAVE",
			"HAZE", "HOLE", "CAPE", "HUSE", "LACE", "LINE",
			"LIVE", "LOVE", "MORE", "MOSE", "MORE", "NICE",
			"RAKE", "ROBE", "ROSE", "SISE", "SIZE", "WARE",
			"WAKE", "WISE", "WINE") && e.encodeESuffix(4)) ||
		(e.stringStart("BLAKE", "BRAKE", "BRINE", "CARLE", "CLEVE", "DUNNE",
			"HEDGE", "HOUSE", "JEFFE", "LUNCE", "STOKE", "STONE",
			"THORE", "WEDGE", "WHITE") && e.encodeESuffix(5)) ||
		(e.stringStart("BRIDGE", "CHEESE") && e.encodeESuffix(6)) ||
		(e.stringAt(-5, "CHARLES")) {
		return true
	}

	return false
}

func (e *Encoder) encodeESuffix(at int) bool {
	//E_Silent_Suffix && !E_Pronouncing_Suffix

	if e.idx == at-1 && len(e.in) > at+1 &&
		(e.isVowelAt(-e.idx+at+1) ||
			(e.stringAt(-e.idx+at, "ST", "SL") && len(e.in) > at+2)) {

		// now filter endings that will cause the 'e' to be pronounced

		// e.g. 'bridgewood' - the other vowels will get eaten
		// up so we need to put one in here
		// e.g. 'bridgette'
		// e.g. 'olena'
		// e.g. 'bridget'
		if e.stringAtEnd(-e.idx+at, "T", "R", "TA", "TT", "NA", "NO", "NE",
			"RS", "RE", "LA", "AU", "RO", "RA", "TTE", "LIA", "NOW", "ROS", "RAS",
			"WOOD", "WATER", "WORTH") {
			return false
		}

		return true
	}

	return false
}

// Exceptions where 'E' is pronounced where it
// usually wouldn't be, and also some cases
// where 'LE' transposition rules don't apply
// and the vowel needs to be encoded here
func (e *Encoder) encodeEPronouncedExceptions() bool {
	// greek names e.g. "herakles" or hispanic names e.g. "robles", where 'e' is pronounced, other exceptions
	if (e.idx+1 == e.lastIdx &&
		(e.stringAtEnd(-3, "OCLES", "ACLES", "AKLES") ||
			e.stringStart("INES",
				"LOPES", "ESTES", "GOMES", "NUNES", "ALVES", "ICKES",
				"INNES", "PERES", "WAGES", "NEVES", "BENES", "DONES",
				"CORTES", "CHAVES", "VALDES", "ROBLES", "TORRES", "FLORES", "BORGES",
				"NIEVES", "MONTES", "SOARES", "VALLES", "GEDDES", "ANDRES", "VIAJES",
				"CALLES", "FONTES", "HERMES", "ACEVES", "BATRES", "MATHES",
				"DELORES", "MORALES", "DOLORES", "ANGELES", "ROSALES", "MIRELES", "LINARES",
				"PERALES", "PAREDES", "BRIONES", "SANCHES", "CAZARES", "REVELES", "ESTEVES",
				"ALVARES", "MATTHES", "SOLARES", "CASARES", "CACERES", "STURGES", "RAMIRES",
				"FUNCHES", "BENITES", "FUENTES", "PUENTES", "TABARES", "HENTGES", "VALORES",
				"GONZALES", "MERCEDES", "FAGUNDES", "JOHANNES", "GONSALES", "BERMUDES",
				"CESPEDES", "BETANCES", "TERRONES", "DIOGENES", "CORRALES", "CABRALES",
				"MARTINES", "GRAJALES",
				"CERVANTES", "FERNANDES", "GONCALVES", "BENEVIDES", "CIFUENTES", "SIFUENTES",
				"SERVANTES", "HERNANDES", "BENAVIDES",
				"ARCHIMEDES", "CARRIZALES", "MAGALLANES"))) ||
		e.stringAt(-2, "FRED", "DGES", "DRED", "GNES") ||
		e.stringAt(-5, "PROBLEM", "RESPLEN") ||
		e.stringAt(-4, "REPLEN") ||
		e.stringAt(-3, "SPLE") {

		return true
	}

	return false
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
// Functions to identify patterns
//////////////////////////////////////////////////////////////////////////////////////////////////////

// isVowel returns true for vowels in many languages and charactersets.
func isVowel(inChar rune) bool {
	return (inChar == 'A') || (inChar == 'E') || (inChar == 'I') || (inChar == 'O') || (inChar == 'U') || (inChar == 'Y') ||
		(inChar == 'À') || (inChar == 'Á') || (inChar == 'Â') || (inChar == 'Ã') || (inChar == 'Ä') || (inChar == 'Å') || (inChar == 'Æ') ||
		(inChar == 'È') || (inChar == 'É') || (inChar == 'Ê') || (inChar == 'Ë') ||
		(inChar == 'Ì') || (inChar == 'Í') || (inChar == 'Î') || (inChar == 'Ï') ||
		(inChar == 'Ò') || (inChar == 'Ó') || (inChar == 'Ô') || (inChar == 'Õ') || (inChar == 'Ö') || (inChar == 'Ø') ||
		(inChar == 'Ù') || (inChar == 'Ú') || (inChar == 'Û') || (inChar == 'Ü') || (inChar == 'Ý') ||
		(inChar == '\uC29F') || (inChar == '\uC28C')
}

/**
 * Tests whether the word is the root or a regular english inflection
 * of it, e.g. "ache", "achy", "aches", "ached", "aching", "achingly"
 * This is for cases where we want to match only the root and corresponding
 * inflected forms, and not completely different words which may have the
 * same substring in them.
 */
func rootOrInflections(inWord []rune, root string) bool {
	lenDiff := len(inWord) - len(root)

	// there's no alternate shorter than the root itself
	if lenDiff < 0 {
		return false
	}

	// inWord must start with all the letters of root except the last
	last := len(root) - 1
	for i := 0; i < last; i++ {
		if inWord[i] != rune(root[i]) {
			return false
		}
	}

	inWord = inWord[last:]
	// at this point we know they start the same way
	// except the last rune of root that we didn't check, so check that now
	// check our last letter and simple plural

	if inWord[0] == rune(root[last]) {
		// last root letter matches
		if lenDiff == 0 {
			//exact match
			return true
		} else if lenDiff == 1 && inWord[1] == 'S' {
			// match with an extra S
			return true
		}
	}

	// different paths if the last letter is 'E' or not
	if root[last] == 'E' {
		// check ED
		if lenDiff == 1 && inWord[0] == 'E' && inWord[1] == 'D' {
			return true
		}
	} else {
		// check +ES
		// check +ED
		// the last character must match if the root doesn't end in E
		if inWord[0] != rune(root[last]) {
			return false
		}

		if lenDiff == 2 &&
			inWord[1] == 'E' && (inWord[2] == 'S' || inWord[2] == 'D') {
			return true
		}
	}

	// at this point our root and inWord match, so now we're just checking the endings
	// of the inWord starting at index "last"

	if lenDiff == 3 && areEqual(inWord, []rune("ING")) {
		// check ING
		return true
	} else if lenDiff == 5 && areEqual(inWord, []rune("INGLY")) {
		// check INGLY
		return true
	} else if lenDiff == 1 && inWord[0] == 'Y' {
		// check Y
		return true
	}

	return false
}

func (e *Encoder) isSlavoGermanic() bool {
	return e.stringStart("SCH", "SW") || e.in[0] == 'J' || e.in[0] == 'W'
}

func (e *Encoder) charNextIs(c rune) bool {
	return e.charAt(1, c)
}

func (e *Encoder) isVowelAt(offset int) bool {
	at := e.idx + offset
	if at < 0 || at >= len(e.in) {
		return false
	}

	return isVowel(e.in[at])
}

func (e *Encoder) charAt(offset int, c rune) bool {
	idx := e.idx + offset
	if idx >= len(e.in) {
		return false
	}

	return e.in[idx] == c
}

// stringAtStart returns true if the given offset is the start of the string and it's starts
// with one of the given vals.  The vals must be given in order of length, shortest to longest, all caps.
func (e *Encoder) stringAtStart(offset int, vals ...string) bool {
	if offset != -e.idx {
		return false
	}
	return e.stringAt(offset, vals...)
}

// stringAtEnd returns true if one of the given substrings is located at the
// relative offset (relative to current idx) given and uses all the remaining
// letters of the input.  The vals must be given in order of length, shortest to longest, all caps.
func (e *Encoder) stringAtEnd(offset int, vals ...string) bool {
	start := e.idx + offset

	// basic bounds check on our start plus
	// if our shortest input would make us run out of chars then none of our inputs could match
	if start < 0 || start >= len(e.in) || start+len(vals[0]) > len(e.in) {
		return false
	}

	// each value given
nextVal:
	for _, v := range vals {
		// bounds check - if we overrun then we know the rest of the list is too long
		// so we're done
		if last, inlen := start+len(v), len(e.in); last > inlen {
			return false
		} else if last < inlen {
			// if we don't land on exactly the end of the input string then we don't need
			// to check the letters
			continue nextVal
		}

		// each letter of the value given
		i := 0
		for _, c := range v {
			if c != e.in[start+i] {
				// char mis-match, this word is done
				continue nextVal
			}
			i++
		}

		// if we make it here we matched all letters
		return true
	}

	// if we make it here we've tried all vals and failed
	return false
}

// stringAt returns true if one of the given substrings is located at the
// relative offset (relative to current idx) given.  The vals must be given in order of
// length, shortest to longest, all caps.
func (e *Encoder) stringAt(offset int, vals ...string) bool {
	start := e.idx + offset

	// basic bounds check on our start plus
	// if our shortest input would make us run out of chars then none of our inputs could match
	if start < 0 || start >= len(e.in) || start+len(vals[0]) > len(e.in) {
		return false
	}

	// each value given
nextVal:
	for _, v := range vals {
		// bounds check - if we fail then we know the rest of the list is too long
		// so we're done
		if start+len(v) > len(e.in) {
			return false
		}

		// each letter of the value given
		i := 0
		for _, c := range v {
			if c != e.in[start+i] {
				// char mis-match, this word is done
				continue nextVal
			}
			i++
		}

		// if we make it here we matched all letters
		return true
	}

	// if we make it here we've tried all vals and failed
	return false
}

func (e *Encoder) stringStart(vals ...string) bool {
	return e.stringAt(-e.idx, vals...)
}

// check each val to see if our string ends with it
// regardless of our current position.  Vals need to be ordered by length, shortest to longest.
func (e *Encoder) stringEnd(vals ...string) bool {

	//each value given
nextVal:
	for _, v := range vals {
		idx := len(e.in) - len(v)
		if idx < 0 {
			// can't match a string that's longer than our input
			return false
		}

		for _, c := range v {
			if c != e.in[idx] {
				// char mis-match, this word is done
				continue nextVal
			}
			idx++
		}

		// if we make it here we matched all letters
		return true
	}

	// if we make it here we've tried all vals and failed
	return false
}

func (e *Encoder) stringExact(vals ...string) bool {
	// each value given
nextVal:
	for _, v := range vals {
		// bounds check - if we fail then we know the rest of the list is too long
		// so we're done
		if len(v) > len(e.in) {
			return false
		} else if len(v) < len(e.in) {
			// too short, next option
			continue nextVal
		}

		i := 0
		for _, c := range v {
			if c != e.in[i] {
				// char mis-match, this word is done
				continue nextVal
			}
			i++
		}

		// if we make it here we matched all letters
		return true
	}

	// if we make it here we've tried all vals and failed
	return false
}

// val must be all caps and not be blank
func (e *Encoder) stringContains(val string) bool {

	lastPossibleStart := len(e.in) - len(val)
	// simple bounds check
	if lastPossibleStart < 0 {
		return false
	}

	// each letter in our input
	for i := 0; i <= lastPossibleStart; i++ {
		// check that against our search val
		// keep checking until we run out of letters in our search val

		tmp := i
		for _, c := range val {
			if e.in[tmp] != c {
				// we're done with this iteration
				break
			}
			// next character in our input
			tmp++
		}
		// if we matched all our search val then we're done
		if tmp-i == len(val) {
			return true
		}
	}

	return false
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
// Functions to mutate the outputs
//////////////////////////////////////////////////////////////////////////////////////////////////////

// Adds encoding character to the encoded string (primary and secondary)
func (e *Encoder) metaphAdd(in rune) {
	e.metaphAddAlt(in, in)
}

// Adds given encoding characters to the associated encoded strings
func (e *Encoder) metaphAddAlt(prim, second rune) {
	if prim != unicode.ReplacementChar {
		// don't dupe added A's
		if !(prim == 'A' && len(e.primBuf) > 0 && e.primBuf[len(e.primBuf)-1] == 'A') {
			if debug {
				fmt.Printf("Append Prim: %v at %v\n", string(prim), string(e.in[0:e.idx+1]))
			}
			e.primBuf = append(e.primBuf, prim)
		}
	}

	if second != unicode.ReplacementChar {
		// don't dupe added A's
		if !(second == 'A' && len(e.secondBuf) > 0 && e.secondBuf[len(e.secondBuf)-1] == 'A') {
			if debug {
				fmt.Printf("Append Alt: %v at %v\n", string(second), string(e.in[0:e.idx+1]))
			}
			e.secondBuf = append(e.secondBuf, second)
		}
	}
}

// Adds given strings to the associated encoded strings
func (e *Encoder) metaphAddStr(prim, second string) {
	// don't dupe added A's
	if !(prim == "A" && len(e.primBuf) > 0 && e.primBuf[len(e.primBuf)-1] == 'A') {
		if debug {
			fmt.Printf("Append Prim: %v at %v\n", prim, string(e.in[0:e.idx+1]))
		}
		e.primBuf = append(e.primBuf, []rune(prim)...)
	}

	// don't dupe added A's
	if second != "" && !(second == "A" && len(e.secondBuf) > 0 && e.secondBuf[len(e.secondBuf)-1] == 'A') {
		if debug {
			fmt.Printf("Append Alt: %v at %v\n", second, string(e.in[0:e.idx+1]))
		}
		e.secondBuf = append(e.secondBuf, []rune(second)...)
	}
}

func (e *Encoder) metaphAddExactApproxAlt(exact, altExact, main, alt string) {
	if e.EncodeExact {
		e.metaphAddStr(exact, altExact)
	} else {
		e.metaphAddStr(main, alt)
	}
}

func (e *Encoder) metaphAddExactApprox(exact, main string) {
	if e.EncodeExact {
		e.metaphAddStr(exact, exact)
	} else {
		e.metaphAddStr(main, main)
	}
}

func (e *Encoder) skipVowels(at int) int {
	if at < 0 {
		return 0
	}
	if at >= len(e.in) {
		return len(e.in)
	}

	it := e.in[at]
	off := at - e.idx

	for isVowel(it) || it == 'W' {

		if e.stringAt(off, "WICZ", "WITZ", "WIAK") ||
			e.stringAt(off-1, "EWSKI", "EWSKY", "OWSKI", "OWSKY") ||
			e.stringAtEnd(off, "WICKI", "WACKI") {
			break
		}

		off++
		if e.charAt(off-1, 'W') &&
			e.charAt(off, 'H') &&
			!e.stringAt(off, "HOP", "HIDE", "HARD", "HEAD", "HAWK", "HERD", "HOOK", "HAND", "HOLE",
				"HEART", "HOUSE", "HOUND", "HAMMER") {

			off++
		}

		if e.idx+off > e.lastIdx {
			break
		}

		it = e.in[e.idx+off]
	}

	if off < 1 {
		panic("bug: skipping vowels moving backward")
	}

	return e.idx + off - 1
}

func (e *Encoder) advanceCounter(noEncodeVowel, encodeVowel int) {
	if e.EncodeVowels {
		e.idx += encodeVowel
	} else {
		e.idx += noEncodeVowel
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
// Misc helper functions
//////////////////////////////////////////////////////////////////////////////////////////////////////

func areEqual(buf1 []rune, buf2 []rune) bool {
	if len(buf1) != len(buf2) {
		return false
	}

	for i := 0; i < len(buf1); i++ {
		if buf1[i] != buf2[i] {
			return false
		}
	}

	return true
}

// make sure we have capacity for our whole buffer, but 0 len
func primeBuf(buf []rune, ensureCap int) []rune {
	if want := ensureCap - cap(buf); want > 0 {
		buf = make([]rune, 0, ensureCap)
	} else if len(buf) != 0 {
		buf = buf[0:0]
	}

	return buf
}
