package mp4

const (
	parentAtom = iota
	simpleParentAtom
	dualAtom
	childAtom
	unknownAtomType
)

const (
	requiredOnePerFile = iota
	requiredOnePerContainer
	requiredVariable
	dependsOnParent
	optionalOnePerFile
	optionalOnePerContainer
	optionalMany
	unknownRequirements
)

const (
	simpleAtom = iota
	versionedAtom
	extendedAtom
	packedLangAtom
	unknownAtom
)

type atomDef struct {
	parents      []string
	container    int
	requirements int
	boxType      int
}

var atomDefs = map[string]atomDef{
	"ftyp": {{"FILE_LEVEL"}, childAtom, requiredOnePerFile, simpleAtom},

	"moov": {{"FILE_LEVEL"}, parentAtom, requiredOnePerFile, simpleAtom},

	"mdat": {{"FILE_LEVEL"}, childAtom, optionalMany, simpleAtom},

	"pdin": {{"FILE_LEVEL"}, childAtom, optionalOnePerFile, versionedAtom},

	"moof": {{"FILE_LEVEL"}, parentAtom, optionalMany, simpleAtom},
	"mfhd": {{"moof"}, childAtom, requiredOnePerContainer, versionedAtom},
	"traf": {{"moof"}, parentAtom, optionalOnePerContainer, simpleAtom},
	"tfhd": {{"traf"}, childAtom, requiredOnePerContainer, versionedAtom},
	"trun": {{"traf"}, childAtom, requiredOnePerContainer, versionedAtom},

	"mfra": {{"FILE_LEVEL"}, parentAtom, optionalOnePerFile, simpleAtom},
	"tfra": {{"mfra"}, childAtom, optionalOnePerContainer, versionedAtom},
	"mfro": {{"mfra"}, childAtom, requiredOnePerContainer, versionedAtom},

	"free": {{"_ANY_LEVEL"}, childAtom, optionalMany, simpleAtom},
	"skip": {{"_ANY_LEVEL"}, childAtom, optionalMany, simpleAtom},

	"uuid": {{"_ANY_LEVEL"}, childAtom, requiredOnePerFile, EXTENDED_ATOM},

	"mvhd": {{"moov"}, childAtom, requiredOnePerFile, versionedAtom},
	"iods": {{"moov"}, childAtom, optionalOnePerFile, versionedAtom},
	// 3gp/MobileMP4
	"drm ": {{"moov"}, childAtom, optionalOnePerFile, versionedAtom},
	"trak": {{"moov"}, parentAtom, optionalMany, simpleAtom},

	"tkhd": {{"trak"}, childAtom, optionalMany, versionedAtom},
	"tref": {{"trak"}, parentAtom, optionalMany, simpleAtom},
	"mdia": {{"trak"}, parentAtom, optionalOnePerContainer, simpleAtom},

	"tapt": {{"trak"}, parentAtom, optionalOnePerContainer, simpleAtom},
	"clef": {{"tapt"}, childAtom, optionalOnePerContainer, versionedAtom},
	"prof": {{"tapt"}, childAtom, optionalOnePerContainer, versionedAtom},
	"enof": {{"tapt"}, childAtom, optionalOnePerContainer, versionedAtom},

	"mdhd": {{"mdia"}, childAtom, optionalOnePerContainer, versionedAtom},
	"minf": {{"mdia"}, parentAtom, requiredOnePerContainer, simpleAtom},

	//minf parent present in chapterized
	"hdlr": {{"mdia", "meta", "minf"}, childAtom, requiredOnePerContainer, versionedAtom},

	"vmhd": {{"minf"}, childAtom, REQ_FAMILIAL_ONE, versionedAtom},
	"smhd": {{"minf"}, childAtom, REQ_FAMILIAL_ONE, versionedAtom},
	"hmhd": {{"minf"}, childAtom, REQ_FAMILIAL_ONE, versionedAtom},
	"nmhd": {{"minf"}, childAtom, REQ_FAMILIAL_ONE, versionedAtom},
	//present in chapterized
	"gmhd": {{"minf"}, childAtom, REQ_FAMILIAL_ONE, versionedAtom},

	//required in minf
	"dinf": {{"minf", "meta"}, parentAtom, optionalOnePerContainer, simpleAtom},

	"url ": {{"dinf"}, childAtom, REQ_FAMILIAL_ONE, versionedAtom},
	"urn ": {{"dinf"}, childAtom, REQ_FAMILIAL_ONE, versionedAtom},
	"dref": {{"dinf"}, childAtom, REQ_FAMILIAL_ONE, versionedAtom},

	"stbl": {{"minf"}, parentAtom, requiredOnePerContainer, simpleAtom},
	"stts": {{"stbl"}, childAtom, requiredOnePerContainer, versionedAtom},
	"ctts": {{"stbl"}, childAtom, optionalOnePerContainer, versionedAtom},
	"stsd": {{"stbl"}, DUAL_STATE_ATOM, requiredOnePerContainer, versionedAtom},

	"stsz": {{"stbl"}, childAtom, REQ_FAMILIAL_ONE, versionedAtom},
	"stz2": {{"stbl"}, childAtom, REQ_FAMILIAL_ONE, versionedAtom},

	"stsc": {{"stbl"}, childAtom, requiredOnePerContainer, versionedAtom},

	"stco": {{"stbl"}, childAtom, REQ_FAMILIAL_ONE, versionedAtom},
	"co64": {{"stbl"}, childAtom, REQ_FAMILIAL_ONE, versionedAtom},

	"stss": {{"stbl"}, childAtom, optionalOnePerContainer, versionedAtom},
	"stsh": {{"stbl"}, childAtom, optionalOnePerContainer, versionedAtom},
	"stdp": {{"stbl"}, childAtom, optionalOnePerContainer, versionedAtom},
	"padb": {{"stbl"}, childAtom, optionalOnePerContainer, versionedAtom},
	"sdtp": {{"stbl", "traf"}, childAtom, optionalOnePerContainer, versionedAtom},
	"sbgp": {{"stbl", "traf"}, childAtom, optionalMany, versionedAtom},
	"sbgp": {{"stbl"}, childAtom, optionalMany, versionedAtom},
	"stps": {{"stbl"}, childAtom, optionalOnePerContainer, versionedAtom},

	"edts": {{"trak"}, parentAtom, optionalOnePerContainer, simpleAtom},
	"elst": {{"edts"}, childAtom, optionalOnePerContainer, versionedAtom},

	"udta": {{"moov", "trak"}, parentAtom, optionalOnePerContainer, simpleAtom},

	//optionally contains info
	"meta": {{"FILE_LEVEL", "moov", "trak", "udta"}, DUAL_STATE_ATOM, optionalOnePerContainer, versionedAtom},

	"mvex": {{"moov"}, parentAtom, optionalOnePerFile, simpleAtom},
	"mehd": {{"mvex"}, childAtom, optionalOnePerFile, versionedAtom},
	"trex": {{"mvex"}, childAtom, requiredOnePerContainer, versionedAtom},

	//"stsl": {	{"????"},						childAtom,				optionalOnePerContainer,					versionedAtom },				//contained by a sample entry box

	"subs": {{"stbl", "traf"}, childAtom, optionalOnePerContainer, versionedAtom},

	"xml ": {{"meta"}, childAtom, optionalOnePerContainer, versionedAtom},
	"bxml": {{"meta"}, childAtom, optionalOnePerContainer, versionedAtom},
	"iloc": {{"meta"}, childAtom, optionalOnePerContainer, versionedAtom},
	"pitm": {{"meta"}, childAtom, optionalOnePerContainer, versionedAtom},
	"ipro": {{"meta"}, parentAtom, optionalOnePerContainer, versionedAtom},
	"infe": {{"meta"}, childAtom, optionalOnePerContainer, versionedAtom},
	"iinf": {{"meta"}, childAtom, optionalOnePerContainer, versionedAtom},

	//parent atom is also "Protected Sample Entry"
	"sinf": {{"ipro", "drms", "drmi"}, parentAtom, requiredOnePerContainer, simpleAtom},
	"frma": {{"sinf"}, childAtom, requiredOnePerContainer, simpleAtom},
	"imif": {{"sinf"}, childAtom, optionalOnePerContainer, versionedAtom},
	"schm": {{"sinf", "srpp"}, childAtom, optionalOnePerContainer, versionedAtom},
	"schi": {{"sinf", "srpp"}, DUAL_STATE_ATOM, optionalOnePerContainer, simpleAtom},
	"skcr": {{"sinf"}, childAtom, optionalOnePerContainer, versionedAtom},

	"user": {{"schi"}, childAtom, optionalOnePerContainer, simpleAtom},
	//could be required in 'drms'/'drmi'
	"key ": {{"schi"}, childAtom, optionalOnePerContainer, versionedAtom},
	"iviv": {{"schi"}, childAtom, optionalOnePerContainer, simpleAtom},
	"righ": {{"schi"}, childAtom, optionalOnePerContainer, simpleAtom},
	"name": {{"schi"}, childAtom, optionalOnePerContainer, simpleAtom},
	"priv": {{"schi"}, childAtom, optionalOnePerContainer, simpleAtom},

	// 'iAEC', '264b', 'iOMA', 'ICSD'
	"iKMS": {{"schi"}, childAtom, optionalOnePerContainer, versionedAtom},
	"iSFM": {{"schi"}, childAtom, optionalOnePerContainer, versionedAtom},
	//boxes with 'k***' are also here; reserved
	"iSLT": {{"schi"}, childAtom, optionalOnePerContainer, simpleAtom},
	"IKEY": {{"tref"}, childAtom, optionalOnePerContainer, simpleAtom},
	"hint": {{"tref"}, childAtom, optionalOnePerContainer, simpleAtom},
	"dpnd": {{"tref"}, childAtom, optionalOnePerContainer, simpleAtom},
	"ipir": {{"tref"}, childAtom, optionalOnePerContainer, simpleAtom},
	"mpod": {{"tref"}, childAtom, optionalOnePerContainer, simpleAtom},
	"sync": {{"tref"}, childAtom, optionalOnePerContainer, simpleAtom},
	//?possible versioned?
	"chap": {{"tref"}, childAtom, optionalOnePerContainer, simpleAtom},

	"ipmc": {{"moov", "meta"}, childAtom, optionalOnePerContainer, versionedAtom},

	"tims": {{"rtp "}, childAtom, requiredOnePerContainer, simpleAtom},
	"tsro": {{"rtp "}, childAtom, optionalOnePerContainer, simpleAtom},
	"snro": {{"rtp "}, childAtom, optionalOnePerContainer, simpleAtom},

	"srpp": {{"srtp"}, childAtom, requiredOnePerContainer, versionedAtom},

	"hnti": {{"udta"}, parentAtom, optionalOnePerContainer, simpleAtom},
	//'rtp ' is defined twice in different containers
	"rtp ": {{"hnti"}, childAtom, optionalOnePerContainer, simpleAtom},
	"sdp ": {{"hnti"}, childAtom, optionalOnePerContainer, simpleAtom},

	"hinf": {{"udta"}, parentAtom, optionalOnePerContainer, simpleAtom},
	"name": {{"udta"}, childAtom, optionalOnePerContainer, simpleAtom},
	"trpy": {{"hinf"}, childAtom, optionalOnePerContainer, simpleAtom},
	"nump": {{"hinf"}, childAtom, optionalOnePerContainer, simpleAtom},
	"tpyl": {{"hinf"}, childAtom, optionalOnePerContainer, simpleAtom},
	"totl": {{"hinf"}, childAtom, optionalOnePerContainer, simpleAtom},
	"npck": {{"hinf"}, childAtom, optionalOnePerContainer, simpleAtom},
	"maxr": {{"hinf"}, childAtom, optionalMany, simpleAtom},
	"dmed": {{"hinf"}, childAtom, optionalOnePerContainer, simpleAtom},
	"dimm": {{"hinf"}, childAtom, optionalOnePerContainer, simpleAtom},
	"drep": {{"hinf"}, childAtom, optionalOnePerContainer, simpleAtom},
	"tmin": {{"hinf"}, childAtom, optionalOnePerContainer, simpleAtom},
	"tmax": {{"hinf"}, childAtom, optionalOnePerContainer, simpleAtom},
	"pmax": {{"hinf"}, childAtom, optionalOnePerContainer, simpleAtom},
	"dmax": {{"hinf"}, childAtom, optionalOnePerContainer, simpleAtom},
	"payt": {{"hinf"}, childAtom, optionalOnePerContainer, simpleAtom},
	"tpay": {{"hinf"}, childAtom, optionalOnePerContainer, simpleAtom},

	"drms": {{"stsd"}, DUAL_STATE_ATOM, REQ_FAMILIAL_ONE, versionedAtom},
	"drmi": {{"stsd"}, DUAL_STATE_ATOM, REQ_FAMILIAL_ONE, versionedAtom},
	"alac": {{"stsd"}, DUAL_STATE_ATOM, REQ_FAMILIAL_ONE, versionedAtom},
	"mp4a": {{"stsd"}, DUAL_STATE_ATOM, REQ_FAMILIAL_ONE, versionedAtom},
	"mp4s": {{"stsd"}, DUAL_STATE_ATOM, REQ_FAMILIAL_ONE, versionedAtom},
	"mp4v": {{"stsd"}, DUAL_STATE_ATOM, REQ_FAMILIAL_ONE, versionedAtom},
	"avc1": {{"stsd"}, DUAL_STATE_ATOM, REQ_FAMILIAL_ONE, versionedAtom},
	"avcp": {{"stsd"}, DUAL_STATE_ATOM, REQ_FAMILIAL_ONE, versionedAtom},
	"text": {{"stsd"}, DUAL_STATE_ATOM, REQ_FAMILIAL_ONE, versionedAtom},
	"jpeg": {{"stsd"}, DUAL_STATE_ATOM, REQ_FAMILIAL_ONE, versionedAtom},
	"tx3g": {{"stsd"}, DUAL_STATE_ATOM, REQ_FAMILIAL_ONE, versionedAtom},
	//"rtp " occurs twice; disparate meanings
	"rtp ": {{"stsd"}, DUAL_STATE_ATOM, REQ_FAMILIAL_ONE, versionedAtom},
	"srtp": {{"stsd"}, DUAL_STATE_ATOM, REQ_FAMILIAL_ONE, simpleAtom},
	"enca": {{"stsd"}, DUAL_STATE_ATOM, REQ_FAMILIAL_ONE, versionedAtom},
	"encv": {{"stsd"}, DUAL_STATE_ATOM, REQ_FAMILIAL_ONE, versionedAtom},
	"enct": {{"stsd"}, DUAL_STATE_ATOM, REQ_FAMILIAL_ONE, versionedAtom},
	"encs": {{"stsd"}, DUAL_STATE_ATOM, REQ_FAMILIAL_ONE, versionedAtom},
	"samr": {{"stsd"}, DUAL_STATE_ATOM, REQ_FAMILIAL_ONE, versionedAtom},
	"sawb": {{"stsd"}, DUAL_STATE_ATOM, REQ_FAMILIAL_ONE, versionedAtom},
	"sawp": {{"stsd"}, DUAL_STATE_ATOM, REQ_FAMILIAL_ONE, versionedAtom},
	"s263": {{"stsd"}, DUAL_STATE_ATOM, REQ_FAMILIAL_ONE, versionedAtom},
	"sevc": {{"stsd"}, DUAL_STATE_ATOM, REQ_FAMILIAL_ONE, versionedAtom},
	"sqcp": {{"stsd"}, DUAL_STATE_ATOM, REQ_FAMILIAL_ONE, versionedAtom},
	"ssmv": {{"stsd"}, DUAL_STATE_ATOM, REQ_FAMILIAL_ONE, versionedAtom},
	"tmcd": {{"stsd"}, DUAL_STATE_ATOM, REQ_FAMILIAL_ONE, versionedAtom},

	"alac": {{"alac"}, childAtom, requiredOnePerContainer, simpleAtom},
	"avcC": {{"avc1", "drmi"}, childAtom, requiredOnePerContainer, simpleAtom},
	"damr": {{"samr", "sawb"}, childAtom, requiredOnePerContainer, simpleAtom},
	"d263": {{"s263"}, childAtom, requiredOnePerContainer, simpleAtom},
	"dawp": {{"sawp"}, childAtom, requiredOnePerContainer, simpleAtom},
	"devc": {{"sevc"}, childAtom, requiredOnePerContainer, simpleAtom},
	"dqcp": {{"sqcp"}, childAtom, requiredOnePerContainer, simpleAtom},
	"dsmv": {{"ssmv"}, childAtom, requiredOnePerContainer, simpleAtom},
	"bitr": {{"d263"}, childAtom, requiredOnePerContainer, simpleAtom},
	//found in NeroAVC
	"btrt": {{"avc1"}, childAtom, optionalOnePerContainer, simpleAtom},
	//?possible versioned?
	"m4ds": {{"avc1"}, childAtom, optionalOnePerContainer, simpleAtom},
	"ftab": {{"tx3g"}, childAtom, optionalOnePerContainer, simpleAtom},

	//the only ISO defined metadata tag; also a 3gp asset
	"cprt": {{"udta"}, childAtom, optionalMany, packedLangAtom},
	//3gp assets
	"titl": {{"udta"}, childAtom, optionalMany, packedLangAtom},
	"auth": {{"udta"}, childAtom, optionalMany, packedLangAtom},
	"perf": {{"udta"}, childAtom, optionalMany, packedLangAtom},
	"gnre": {{"udta"}, childAtom, optionalMany, packedLangAtom},
	"dscp": {{"udta"}, childAtom, optionalMany, packedLangAtom},
	"albm": {{"udta"}, childAtom, optionalMany, packedLangAtom},
	"yrrc": {{"udta"}, childAtom, optionalMany, versionedAtom},
	"rtng": {{"udta"}, childAtom, optionalMany, packedLangAtom},
	"clsf": {{"udta"}, childAtom, optionalMany, packedLangAtom},
	"kywd": {{"udta"}, childAtom, optionalMany, packedLangAtom},
	"loci": {{"udta"}, childAtom, optionalMany, packedLangAtom},

	//id3v2 tag
	"ID32": {{"meta"}, childAtom, optionalMany, packedLangAtom},

	//"chpl": {	{"udta"},						childAtom,				optionalOnePerFile,				versionedAtom },		//Nero - seems to be versioned

	//"ndrm": {	{"udta"},						childAtom,				optionalOnePerFile,				versionedAtom },		//Nero - seems to be versioned

	//"tags": {	{"udta"},						childAtom,				optionalOnePerFile,				simpleAtom },			//Another Nero-Creationª

	// ...so if they claim that "tags doesn't have any children": {
	// why does nerotags.exe say "tshd atom"? If 'tags' doesn't
	// have any children: { then tshd can't be an atom....
	// Clearly: { they are EternallyRightª and everyone else is
	// always wrong.

	//Pish! Seems that Nero is simply unable to register any atoms.

	//iTunes metadata container
	"ilst": {{"meta"}, parentAtom, optionalOnePerFile, simpleAtom},
	//reverse dns metadata
	"----": {{"ilst"}, parentAtom, optionalMany, simpleAtom},
	"mean": {{"----"}, childAtom, requiredOnePerContainer, versionedAtom},
	"name": {{"----"}, childAtom, requiredOnePerContainer, versionedAtom},

	//multiple parents; keep 3rd from end; manual return
	"esds": {{"SAMPLE_DESC"}, childAtom, requiredOnePerContainer, simpleAtom},

	//multiple parents; keep 2nd from end; manual return
	"(..)": {{"ilst"}, parentAtom, optionalOnePerContainer, simpleAtom},
	//multiple parents
	"data": {{"ITUNES_METADATA"}, childAtom, parentSpecific, versionedAtom},
}
