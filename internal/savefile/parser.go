package savefile

import (
	"encoding/binary"
	"errors"
)

// PokemonStats represents a Pokemon's battle stats
type PokemonStats struct {
	HP      int `json:"hp"`
	Attack  int `json:"attack"`
	Defense int `json:"defense"`
	SpAtk   int `json:"spAtk"`
	SpDef   int `json:"spDef"`
	Speed   int `json:"speed"`
}

// PartyPokemon represents a Pokemon in the party
type PartyPokemon struct {
	Species     string       `json:"species"`
	Nickname    string       `json:"nickname"`
	Level       int          `json:"level"`
	SpeciesNum  int          `json:"speciesNum"`
	Nature      string       `json:"nature"`
	Item        string       `json:"item"`
	ItemNum     int          `json:"itemNum"`
	Moves       []string     `json:"moves"`
	MoveNums    []int        `json:"moveNums"`
	Stats       PokemonStats `json:"stats"`
	CurrentHP   int          `json:"currentHp"`
	AbilitySlot int          `json:"abilitySlot"` // 0 = first ability, 1 = second ability
}

// ParseResult contains the parsed save data
type ParseResult struct {
	Party []PartyPokemon `json:"party"`
}

// Substructure order lookup table (personality % 24)
// Based on Bulbapedia: https://bulbapedia.bulbagarden.net/wiki/Pokémon_data_substructures_(Generation_III)
// Each entry shows [pos0, pos1, pos2, pos3] where value is substructure type (G=0, A=1, E=2, M=3)
var substructOrder = [24][4]int{
	{0, 1, 2, 3}, // 00 GAEM
	{0, 1, 3, 2}, // 01 GAME
	{0, 2, 1, 3}, // 02 GEAM
	{0, 2, 3, 1}, // 03 GEMA
	{0, 3, 1, 2}, // 04 GMAE
	{0, 3, 2, 1}, // 05 GMEA
	{1, 0, 2, 3}, // 06 AGEM
	{1, 0, 3, 2}, // 07 AGME
	{1, 2, 0, 3}, // 08 AEGM
	{1, 2, 3, 0}, // 09 AEMG
	{1, 3, 0, 2}, // 10 AMGE
	{1, 3, 2, 0}, // 11 AMEG
	{2, 0, 1, 3}, // 12 EGAM
	{2, 0, 3, 1}, // 13 EGMA
	{2, 1, 0, 3}, // 14 EAGM
	{2, 1, 3, 0}, // 15 EAMG
	{2, 3, 0, 1}, // 16 EMGA
	{2, 3, 1, 0}, // 17 EMAG
	{3, 0, 1, 2}, // 18 MGAE
	{3, 0, 2, 1}, // 19 MGEA
	{3, 1, 0, 2}, // 20 MAGE
	{3, 1, 2, 0}, // 21 MAEG
	{3, 2, 0, 1}, // 22 MEGA
	{3, 2, 1, 0}, // 23 MEAG
}

// expansionToNationalDex maps pokeemerald-expansion internal species IDs to national dex numbers
// Only Gen 9 Pokemon need mapping; Gen 1-8 internal IDs match national dex
var expansionToNationalDex = map[int]int{
	1289: 906, 1290: 907, 1291: 908, 1292: 909, 1293: 910, 1294: 911,
	1295: 912, 1296: 913, 1297: 914, 1298: 915, 1301: 917, 1302: 918,
	1303: 919, 1304: 920, 1305: 921, 1306: 922, 1307: 923, 1308: 924,
	1311: 926, 1312: 927, 1313: 928, 1314: 929, 1315: 930, 1320: 932,
	1321: 933, 1322: 934, 1323: 935, 1324: 936, 1325: 937, 1326: 938,
	1327: 939, 1328: 940, 1329: 941, 1330: 942, 1331: 943, 1332: 944,
	1333: 945, 1334: 946, 1335: 947, 1336: 948, 1337: 949, 1338: 950,
	1339: 951, 1340: 952, 1341: 953, 1342: 954, 1343: 955, 1344: 956,
	1345: 957, 1346: 958, 1347: 959, 1348: 960, 1349: 961, 1350: 962,
	1351: 963, 1354: 965, 1355: 966, 1356: 967, 1357: 968, 1358: 969,
	1359: 970, 1360: 971, 1361: 972, 1362: 973, 1363: 974, 1364: 975,
	1365: 976, 1366: 977, 1370: 979, 1371: 980, 1372: 981, 1375: 983,
	1376: 984, 1377: 985, 1378: 986, 1379: 987, 1380: 988, 1381: 989,
	1382: 990, 1383: 991, 1384: 992, 1385: 993, 1386: 994, 1387: 995,
	1388: 996, 1389: 997, 1390: 998, 1393: 1000, 1394: 1001, 1395: 1002,
	1396: 1003, 1397: 1004, 1398: 1005, 1399: 1006, 1400: 1007, 1401: 1008,
	1406: 1009, 1407: 1010, 1408: 1011, 1413: 1014, 1414: 1015, 1415: 1016,
	1425: 1018, 1426: 1019, 1427: 1020, 1428: 1021, 1429: 1022, 1430: 1023,
	1434: 1025,
}

// natureNames maps personality % 25 to nature name
var natureNames = []string{
	"Hardy", "Lonely", "Brave", "Adamant", "Naughty",
	"Bold", "Docile", "Relaxed", "Impish", "Lax",
	"Timid", "Hasty", "Serious", "Jolly", "Naive",
	"Modest", "Mild", "Quiet", "Bashful", "Rash",
	"Calm", "Gentle", "Sassy", "Careful", "Quirky",
}

// NatureEffect describes what stats a nature affects
type NatureEffect struct {
	Plus  string `json:"plus"`  // Stat increased by 10%
	Minus string `json:"minus"` // Stat decreased by 10%
}

// natureEffects maps nature name to its stat effects
var natureEffects = map[string]NatureEffect{
	"Hardy":   {Plus: "", Minus: ""},           // Neutral
	"Lonely":  {Plus: "Attack", Minus: "Defense"},
	"Brave":   {Plus: "Attack", Minus: "Speed"},
	"Adamant": {Plus: "Attack", Minus: "Sp. Atk"},
	"Naughty": {Plus: "Attack", Minus: "Sp. Def"},
	"Bold":    {Plus: "Defense", Minus: "Attack"},
	"Docile":  {Plus: "", Minus: ""},           // Neutral
	"Relaxed": {Plus: "Defense", Minus: "Speed"},
	"Impish":  {Plus: "Defense", Minus: "Sp. Atk"},
	"Lax":     {Plus: "Defense", Minus: "Sp. Def"},
	"Timid":   {Plus: "Speed", Minus: "Attack"},
	"Hasty":   {Plus: "Speed", Minus: "Defense"},
	"Serious": {Plus: "", Minus: ""},           // Neutral
	"Jolly":   {Plus: "Speed", Minus: "Sp. Atk"},
	"Naive":   {Plus: "Speed", Minus: "Sp. Def"},
	"Modest":  {Plus: "Sp. Atk", Minus: "Attack"},
	"Mild":    {Plus: "Sp. Atk", Minus: "Defense"},
	"Quiet":   {Plus: "Sp. Atk", Minus: "Speed"},
	"Bashful": {Plus: "", Minus: ""},           // Neutral
	"Rash":    {Plus: "Sp. Atk", Minus: "Sp. Def"},
	"Calm":    {Plus: "Sp. Def", Minus: "Attack"},
	"Gentle":  {Plus: "Sp. Def", Minus: "Defense"},
	"Sassy":   {Plus: "Sp. Def", Minus: "Speed"},
	"Careful": {Plus: "Sp. Def", Minus: "Sp. Atk"},
	"Quirky":  {Plus: "", Minus: ""},           // Neutral
}

// GetNatureEffect returns the stat effects for a nature
func GetNatureEffect(nature string) NatureEffect {
	if effect, ok := natureEffects[nature]; ok {
		return effect
	}
	return NatureEffect{}
}


// expansionItemToShowdown maps pokeemerald-expansion item IDs to Showdown item IDs
var expansionItemToShowdown = map[int]int{
	6: 188,
	7: 304,
	8: 303,
	9: 101,
	10: 115,
	11: 494,
	12: 372,
	13: 401,
	14: 266,
	15: 246,
	16: 264,
	17: 294,
	18: 153,
	19: 258,
	20: 137,
	21: 194,
	22: 111,
	23: 425,
	24: 465,
	25: 325,
	26: 661,
	27: 64,
	53: 22,
	133: 696,
	134: 697,
	136: 27,
	154: 379,
	156: 1,
	165: 195,
	166: 102,
	167: 314,
	168: 418,
	169: 72,
	170: 12,
	171: 449,
	172: 85,
	173: 339,
	174: 694,
	175: 695,
	176: 700,
	177: 701,
	178: 702,
	179: 703,
	211: 142,
	212: 529,
	213: 492,
	214: 241,
	215: 693,
	216: 480,
	217: 295,
	218: 439,
	219: 116,
	220: 92,
	221: 711,
	222: 712,
	223: 719,
	224: 720,
	225: 739,
	226: 740,
	227: 108,
	228: 523,
	229: 367,
	230: 119,
	231: 272,
	232: 113,
	233: 385,
	234: 365,
	235: 692,
	236: 691,
	237: 321,
	238: 704,
	239: 705,
	240: 706,
	241: 707,
	242: 708,
	243: 709,
	244: 710,
	250: 146,
	251: 463,
	252: 572,
	253: 282,
	254: 220,
	255: 143,
	256: 516,
	257: 117,
	258: 450,
	259: 291,
	260: 223,
	261: 477,
	262: 464,
	263: 105,
	264: 110,
	265: 225,
	266: 610,
	267: 103,
	268: 442,
	269: 54,
	270: 67,
	271: 676,
	272: 677,
	273: 679,
	274: 678,
	275: 681,
	276: 668,
	277: 670,
	278: 671,
	279: 669,
	280: 680,
	281: 673,
	282: 672,
	283: 674,
	284: 682,
	285: 683,
	286: 675,
	287: 684,
	288: 698,
	289: 699,
	290: 390,
	291: 41,
	292: 608,
	293: 585,
	294: 586,
	295: 583,
	296: 628,
	297: 622,
	298: 579,
	299: 620,
	300: 588,
	301: 592,
	302: 602,
	303: 589,
	304: 577,
	305: 600,
	306: 601,
	307: 580,
	308: 621,
	309: 605,
	310: 590,
	311: 591,
	312: 607,
	313: 613,
	314: 584,
	315: 612,
	316: 587,
	317: 614,
	318: 598,
	319: 578,
	320: 599,
	321: 596,
	322: 619,
	323: 625,
	324: 615,
	325: 582,
	326: 576,
	327: 623,
	328: 627,
	329: 618,
	330: 629,
	331: 630,
	332: 626,
	333: 573,
	334: 594,
	335: 575,
	336: 616,
	337: 617,
	338: 624,
	339: 307,
	340: 141,
	341: 528,
	342: 120,
	343: 172,
	344: 218,
	345: 139,
	346: 344,
	347: 182,
	348: 149,
	349: 369,
	350: 53,
	351: 415,
	352: 161,
	353: 107,
	354: 89,
	355: 473,
	356: 611,
	357: 631,
	358: 632,
	359: 633,
	360: 634,
	361: 635,
	362: 636,
	363: 637,
	364: 638,
	365: 639,
	366: 640,
	367: 641,
	368: 642,
	369: 643,
	370: 644,
	371: 645,
	372: 646,
	373: 647,
	374: 648,
	375: 649,
	376: 657,
	377: 656,
	378: 658,
	379: 650,
	380: 651,
	381: 652,
	382: 689,
	383: 688,
	384: 690,
	385: 653,
	386: 685,
	387: 686,
	388: 654,
	389: 655,
	390: 659,
	391: 687,
	392: 251,
	394: 491,
	395: 261,
	397: 374,
	398: 93,
	399: 94,
	400: 459,
	401: 4,
	402: 265,
	403: 180,
	404: 430,
	405: 240,
	406: 312,
	407: 416,
	408: 155,
	409: 531,
	410: 419,
	418: 269,
	419: 360,
	420: 357,
	421: 356,
	422: 359,
	423: 355,
	424: 354,
	425: 444,
	426: 61,
	427: 300,
	428: 273,
	430: 305,
	431: 32,
	432: 343,
	433: 456,
	434: 436,
	435: 520,
	436: 447,
	437: 187,
	438: 461,
	439: 106,
	440: 35,
	441: 286,
	442: 68,
	443: 70,
	444: 69,
	445: 145,
	446: 515,
	447: 88,
	448: 193,
	449: 453,
	450: 221,
	451: 664,
	452: 665,
	453: 666,
	454: 667,
	455: 2,
	456: 60,
	457: 595,
	458: 606,
	459: 51,
	460: 535,
	464: 285,
	465: 236,
	469: 150,
	471: 429,
	472: 242,
	473: 438,
	474: 537,
	475: 297,
	476: 539,
	477: 132,
	478: 252,
	479: 249,
	481: 151,
	482: 574,
	483: 289,
	484: 224,
	485: 237,
	486: 95,
	487: 34,
	488: 179,
	489: 476,
	490: 437,
	491: 29,
	492: 382,
	493: 383,
	494: 130,
	495: 147,
	496: 417,
	497: 6,
	498: 387,
	499: 410,
	500: 31,
	501: 118,
	502: 609,
	503: 581,
	504: 604,
	505: 660,
	506: 662,
	507: 663,
	508: 713,
	509: 714,
	510: 715,
	511: 716,
	512: 717,
	513: 718,
	514: 63,
	515: 65,
	516: 333,
	517: 381,
	518: 13,
	519: 244,
	520: 319,
	521: 334,
	522: 262,
	523: 448,
	524: 140,
	525: 538,
	526: 274,
	527: 5,
	528: 217,
	529: 384,
	530: 44,
	531: 302,
	532: 533,
	533: 337,
	534: 351,
	535: 235,
	536: 371,
	537: 213,
	538: 178,
	539: 486,
	540: 81,
	541: 275,
	542: 375,
	543: 306,
	544: 462,
	545: 323,
	546: 530,
	547: 114,
	548: 21,
	549: 66,
	550: 311,
	551: 329,
	552: 526,
	553: 409,
	554: 567,
	555: 71,
	556: 234,
	557: 443,
	558: 76,
	559: 330,
	560: 487,
	561: 62,
	562: 233,
	563: 185,
	564: 78,
	565: 17,
	566: 603,
	567: 248,
	568: 158,
	569: 426,
	570: 335,
	571: 10,
	572: 238,
	573: 472,
	574: 124,
	575: 290,
	576: 86,
	577: 230,
	578: 420,
	579: 593,
	580: 597,
	758: 746,
	759: 747,
	760: 749,
	761: 750,
	762: 751,
	763: 753,
	764: 745,
	768: 744,
	769: 748,
	792: 741,
	793: 743,
	794: 742,
	798: 388,
	799: 754,
	800: 755,
	801: 756,
	802: 757,
	803: 758,
	804: 759,
	805: 760,
	814: 761,
	828: 308,
	829: 544,
	830: 545,
	831: 546,
	832: 547,
	833: 548,
	834: 549,
	835: 550,
	836: 551,
	837: 552,
	838: 553,
	839: 554,
	840: 555,
	841: 556,
	842: 557,
	843: 558,
	844: 559,
	845: 560,
	846: 561,
	847: 562,
	848: 563,
	849: 564,
	850: 565,
	851: 566,
	852: 568,
	853: 569,
	854: 570,
}

// Gen 3 character encoding table
var gen3CharTable = map[byte]rune{
	0xBB: 'A', 0xBC: 'B', 0xBD: 'C', 0xBE: 'D', 0xBF: 'E',
	0xC0: 'F', 0xC1: 'G', 0xC2: 'H', 0xC3: 'I', 0xC4: 'J',
	0xC5: 'K', 0xC6: 'L', 0xC7: 'M', 0xC8: 'N', 0xC9: 'O',
	0xCA: 'P', 0xCB: 'Q', 0xCC: 'R', 0xCD: 'S', 0xCE: 'T',
	0xCF: 'U', 0xD0: 'V', 0xD1: 'W', 0xD2: 'X', 0xD3: 'Y',
	0xD4: 'Z',
	0xD5: 'a', 0xD6: 'b', 0xD7: 'c', 0xD8: 'd', 0xD9: 'e',
	0xDA: 'f', 0xDB: 'g', 0xDC: 'h', 0xDD: 'i', 0xDE: 'j',
	0xDF: 'k', 0xE0: 'l', 0xE1: 'm', 0xE2: 'n', 0xE3: 'o',
	0xE4: 'p', 0xE5: 'q', 0xE6: 'r', 0xE7: 's', 0xE8: 't',
	0xE9: 'u', 0xEA: 'v', 0xEB: 'w', 0xEC: 'x', 0xED: 'y',
	0xEE: 'z',
	0xA1: '0', 0xA2: '1', 0xA3: '2', 0xA4: '3', 0xA5: '4',
	0xA6: '5', 0xA7: '6', 0xA8: '7', 0xA9: '8', 0xAA: '9',
	0xAB: '!', 0xAC: '?', 0xAD: '.', 0xAE: '-', 0xB0: '\u2026',
	0xB1: '\u201C', 0xB2: '\u201D', 0xB3: '\u2018', 0xB4: '\u2019', 0xB5: '\u2642',
	0xB6: '\u2640', 0xB8: ',', 0xB9: '/', 0x00: ' ',
	0xFF: 0, // Terminator
}

// ParseGen3Save parses a Gen 3 Pokemon save file and extracts party Pokemon
func ParseGen3Save(data []byte) (*ParseResult, error) {
	if len(data) < 0x20000 {
		return nil, errors.New("save file too small")
	}

	result := &ParseResult{
		Party: []PartyPokemon{},
	}

	// Gen 3 saves use sectors of 0x1000 bytes each
	// There are 14 sectors per save slot (0-13)
	// Save A: sectors 0x0000-0xDFFF, Save B: sectors 0xE000-0x1BFFF
	// Each sector has a footer at offset 0xFF4 with section ID (2 bytes)
	// Section 1 contains party data at offset 0x234 (count) and 0x238 (data)

	// Try Save B first (usually more recent), then Save A
	saveSlots := []int{0xE000, 0x0000}

	for _, slotBase := range saveSlots {
		// Find Section 1 (Team/Items) by checking sector footers
		section1Offset := -1

		for sectorIdx := 0; sectorIdx < 14; sectorIdx++ {
			sectorBase := slotBase + (sectorIdx * 0x1000)
			footerOffset := sectorBase + 0xFF4

			if footerOffset+4 > len(data) {
				continue
			}

			sectionID := binary.LittleEndian.Uint16(data[footerOffset : footerOffset+2])
			if sectionID == 1 {
				section1Offset = sectorBase
				break
			}
		}

		if section1Offset == -1 {
			continue
		}

		partyCountOffset := section1Offset + 0x234
		partyDataOffset := section1Offset + 0x238

		if partyDataOffset+600 > len(data) {
			continue
		}

		partyCount := int(data[partyCountOffset])
		if partyCount < 1 || partyCount > 6 {
			continue
		}

		// Parse each party Pokemon
		for i := 0; i < partyCount; i++ {
			pokemonOffset := partyDataOffset + (i * 100)
			pokemon := parsePokemon(data[pokemonOffset : pokemonOffset+100])
			if pokemon.SpeciesNum > 0 && pokemon.SpeciesNum <= 1025 {
				result.Party = append(result.Party, pokemon)
			}
		}

		if len(result.Party) > 0 {
			break
		}
	}

	if len(result.Party) == 0 {
		return nil, errors.New("no party Pokemon found")
	}

	return result, nil
}

func parsePokemon(data []byte) PartyPokemon {
	if len(data) < 100 {
		return PartyPokemon{}
	}

	// Read personality value (bytes 0-3)
	personality := binary.LittleEndian.Uint32(data[0:4])

	// Read OT ID (bytes 4-7)
	otID := binary.LittleEndian.Uint32(data[4:8])

	// Read base nickname (bytes 8-17, 10 characters)
	// Extended nickname chars (11-12) are in Growth substruct, extracted after decryption
	baseNickname := decodeGen3String(data[8:18])

	// Read level from party data (byte 84)
	level := int(data[84])

	// Calculate nature from personality
	nature := natureNames[personality%25]

	// Decrypt the data block (bytes 32-79)
	encryptionKey := personality ^ otID
	decryptedData := make([]byte, 48)
	copy(decryptedData, data[32:80])

	// XOR each 4-byte word
	for i := 0; i < 48; i += 4 {
		word := binary.LittleEndian.Uint32(decryptedData[i : i+4])
		word ^= encryptionKey
		binary.LittleEndian.PutUint32(decryptedData[i:i+4], word)
	}

	// Find substructure positions
	order := substructOrder[personality%24]
	typeToPos := make(map[int]int)
	for pos, typ := range order {
		typeToPos[typ] = pos
	}

	growthPos := typeToPos[0] * 12  // G = 0
	attacksPos := typeToPos[1] * 12 // A = 1
	miscPos := typeToPos[3] * 12    // M = 3

	// Get species ID from Growth substructure (bytes 0-1)
	rawSpeciesID := int(binary.LittleEndian.Uint16(decryptedData[growthPos : growthPos+2]))
	speciesNum := rawSpeciesID & 0x7FF
	if nationalDex, ok := expansionToNationalDex[speciesNum]; ok {
		speciesNum = nationalDex
	}

	// Get item ID from Growth substructure (bytes 2-3)
	rawItemID := int(binary.LittleEndian.Uint16(decryptedData[growthPos+2 : growthPos+4]))
	itemNum := rawItemID
	if showdownID, ok := expansionItemToShowdown[rawItemID]; ok {
		itemNum = showdownID
	}

	// Get extended nickname characters (11th and 12th) from Growth substruct
	// nickname11: bytes 4-7, bits 21-28
	// nickname12: bytes 10-11, bits 6-13
	expData := binary.LittleEndian.Uint32(decryptedData[growthPos+4 : growthPos+8])
	nickname11 := byte((expData >> 21) & 0xFF)
	pokeballData := binary.LittleEndian.Uint16(decryptedData[growthPos+10 : growthPos+12])
	nickname12 := byte((pokeballData >> 6) & 0xFF)

	// Build full nickname (base + extended chars if present)
	nickname := baseNickname
	if nickname11 != 0xFF && nickname11 != 0 {
		if char, ok := gen3CharTable[nickname11]; ok && char != 0 {
			nickname += string(char)
		}
	}
	if nickname12 != 0xFF && nickname12 != 0 {
		if char, ok := gen3CharTable[nickname12]; ok && char != 0 {
			nickname += string(char)
		}
	}

	// Get moves from Attacks substructure (bytes 0-7, 4 moves of 2 bytes each)
	moveNums := make([]int, 0, 4)
	for i := 0; i < 4; i++ {
		moveID := int(binary.LittleEndian.Uint16(decryptedData[attacksPos+i*2 : attacksPos+i*2+2]))
		if moveID > 0 {
			moveNums = append(moveNums, moveID)
		}
	}

	// Get ability slot from Misc substructure (bits 29-30 of bytes 8-11)
	// In pokeemerald-expansion, abilityNum is a 2-bit field in the ribbon data:
	// Bytes 8-11 contain: ribbons (29 bits) + abilityNum (2 bits) + fatefulEncounter (1 bit)
	ribbonData := binary.LittleEndian.Uint32(decryptedData[miscPos+8 : miscPos+12])
	abilitySlot := int((ribbonData >> 29) & 3) // 2 bits for ability index (0, 1, or 2 for hidden)

	// Get stats from party data section (bytes 86-99)
	// Party data: status(4), level(1), pokerus(1), currentHP(2), maxHP(2), atk(2), def(2), spe(2), spa(2), spd(2)
	currentHP := int(binary.LittleEndian.Uint16(data[86:88]))
	stats := PokemonStats{
		HP:      int(binary.LittleEndian.Uint16(data[88:90])),
		Attack:  int(binary.LittleEndian.Uint16(data[90:92])),
		Defense: int(binary.LittleEndian.Uint16(data[92:94])),
		Speed:   int(binary.LittleEndian.Uint16(data[94:96])),
		SpAtk:   int(binary.LittleEndian.Uint16(data[96:98])),
		SpDef:   int(binary.LittleEndian.Uint16(data[98:100])),
	}

	return PartyPokemon{
		Nickname:    nickname,
		Level:       level,
		SpeciesNum:  speciesNum,
		Nature:      nature,
		ItemNum:     itemNum,
		MoveNums:    moveNums,
		Stats:       stats,
		CurrentHP:   currentHP,
		AbilitySlot: abilitySlot,
	}
}

func decodeGen3String(data []byte) string {
	result := make([]rune, 0, len(data))
	for _, b := range data {
		if b == 0xFF { // String terminator
			break
		}
		if char, ok := gen3CharTable[b]; ok && char != 0 {
			result = append(result, char)
		}
	}
	return string(result)
}
