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
	Species    string       `json:"species"`
	Nickname   string       `json:"nickname"`
	Level      int          `json:"level"`
	SpeciesNum int          `json:"speciesNum"`
	Nature     string       `json:"nature"`
	Item       string       `json:"item"`
	ItemNum    int          `json:"itemNum"`
	Moves      []string     `json:"moves"`
	MoveNums   []int        `json:"moveNums"`
	Stats      PokemonStats `json:"stats"`
	CurrentHP  int          `json:"currentHp"`
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
	6: 188, 7: 304, 8: 303, 9: 101, 10: 115, 11: 494, 12: 372, 13: 401,
	14: 266, 15: 246, 16: 264, 17: 294, 18: 153, 19: 258, 20: 137, 21: 194,
	22: 111, 23: 425, 24: 465, 25: 325, 26: 661, 27: 64, 53: 22, 133: 696,
	134: 697, 136: 27, 154: 379, 156: 1, 165: 195, 166: 102, 167: 314, 168: 418,
	169: 72, 170: 12, 171: 449, 172: 85, 173: 339, 174: 694, 175: 695, 176: 700,
	177: 701, 178: 702, 179: 703, 211: 142, 212: 529, 213: 492, 214: 241, 215: 693,
	216: 480, 217: 295, 218: 439, 219: 116, 220: 92, 221: 711, 222: 712, 223: 719,
	224: 720, 225: 739, 226: 740, 227: 108, 228: 523, 229: 367, 230: 119, 231: 272,
	232: 113, 233: 385, 234: 365, 235: 692, 236: 691, 237: 321, 238: 704, 239: 705,
	240: 706, 241: 707, 242: 708, 243: 709, 244: 710, 250: 146, 251: 463, 252: 572,
	253: 282, 254: 220, 255: 143, 256: 516, 257: 117, 258: 450, 259: 291, 260: 223,
	261: 477, 262: 464, 263: 105, 264: 110, 265: 225, 266: 610, 267: 103, 268: 442,
	269: 54, 270: 89, 271: 289, 272: 297, 273: 253, 274: 309, 275: 298, 276: 276,
	277: 250, 278: 244, 279: 251, 280: 249, 281: 254, 282: 243, 283: 245, 284: 252,
	285: 248, 286: 247, 287: 302, 288: 303, 289: 301, 290: 573, 291: 574, 292: 575,
	293: 576, 294: 577, 295: 578, 296: 579, 297: 580, 298: 581, 299: 582, 300: 583,
	301: 584, 302: 585, 303: 586, 304: 587, 305: 588, 306: 589, 307: 590, 308: 591,
	309: 592, 310: 593, 311: 594, 312: 595, 313: 596, 314: 597, 315: 598, 316: 599,
	317: 600, 318: 601, 319: 602, 320: 603, 321: 604, 322: 605, 323: 606, 324: 607,
	325: 608, 326: 609, 327: 307, 328: 310, 329: 148, 330: 149, 331: 150, 332: 151,
	333: 152, 334: 154, 335: 155, 336: 156, 337: 157, 338: 158, 339: 159, 340: 160,
	341: 161, 342: 162, 343: 172, 344: 163, 345: 164, 346: 165, 347: 166, 348: 167,
	349: 168, 350: 169, 351: 170, 352: 171, 353: 173, 354: 174, 355: 175, 356: 176,
	357: 177, 358: 178, 359: 179, 360: 180, 361: 181, 362: 182, 363: 183, 364: 184,
	365: 185, 366: 186, 367: 187, 376: 189, 377: 190, 378: 191, 379: 192, 380: 193,
	381: 2, 382: 3, 383: 4, 384: 5, 385: 6, 386: 7, 387: 8, 388: 9, 389: 10, 390: 11,
	391: 313, 392: 13, 393: 14, 394: 15, 395: 16, 396: 17, 397: 18, 398: 19, 399: 20,
	400: 21, 404: 430, 405: 240, 406: 312, 407: 416, 408: 155, 409: 531, 410: 419,
	415: 23, 416: 24, 417: 25, 418: 26, 420: 28, 421: 29, 422: 30, 423: 31,
	424: 32, 425: 33, 426: 34, 427: 35, 428: 36, 429: 37, 430: 38, 431: 39, 432: 40,
	433: 41, 434: 42, 435: 43, 436: 44, 437: 45, 438: 46, 439: 47, 440: 48, 441: 49,
	442: 50, 443: 51, 444: 52, 445: 53, 446: 55, 447: 56, 448: 57, 449: 58, 450: 59,
	451: 60, 452: 61, 453: 62, 454: 63, 458: 65, 459: 66, 460: 67, 461: 68, 462: 69,
	463: 70, 464: 71, 467: 73, 468: 74, 469: 75, 470: 76, 471: 77, 472: 78, 473: 79,
	474: 80, 475: 81, 476: 82, 477: 83, 478: 84, 481: 86, 482: 87, 483: 88, 486: 90,
	487: 91, 490: 93, 491: 94, 492: 95, 493: 96, 494: 97, 495: 98, 496: 99, 497: 100,
	500: 104, 503: 106, 504: 107, 507: 109, 510: 112, 513: 114, 516: 118, 519: 120,
	520: 121, 521: 122, 522: 123, 523: 124, 524: 125, 525: 126, 526: 127, 527: 128,
	528: 129, 529: 130, 530: 131, 531: 132, 532: 133, 533: 134, 534: 135, 535: 136,
	538: 138, 539: 139, 540: 140, 541: 141, 544: 144, 545: 145, 548: 147, 580: 196,
	581: 197, 582: 198, 583: 199, 584: 200, 585: 201, 586: 202, 587: 203, 588: 204,
	589: 205, 590: 206, 591: 207, 592: 208, 593: 209, 594: 210, 595: 211, 596: 212,
	597: 213, 598: 214, 599: 215, 600: 216, 601: 217, 602: 218, 603: 219, 606: 221,
	607: 222, 610: 224, 613: 226, 614: 227, 615: 228, 616: 229, 617: 230, 618: 231,
	619: 232, 620: 233, 621: 234, 622: 235, 623: 236, 624: 237, 625: 238, 626: 239,
	627: 240, 630: 242, 651: 255, 652: 256, 653: 257, 656: 259, 657: 260, 658: 261,
	659: 262, 660: 263, 663: 265, 666: 267, 667: 268, 668: 269, 669: 270, 670: 271,
	673: 273, 674: 274, 675: 275, 678: 277, 679: 278, 680: 279, 681: 280, 682: 281,
	685: 283, 686: 284, 687: 285, 688: 286, 689: 287, 690: 288, 693: 290, 696: 292,
	697: 293, 700: 296, 703: 299, 704: 300, 709: 305, 710: 306, 713: 311, 714: 312,
	717: 315, 718: 316, 719: 317, 720: 318, 721: 319, 722: 320, 725: 322, 726: 323,
	727: 324, 730: 326, 731: 327, 732: 328, 733: 329, 734: 330, 735: 331, 736: 332,
	737: 333, 738: 334, 739: 335, 740: 336, 741: 337, 742: 338, 745: 340, 746: 341,
	747: 342, 750: 344, 751: 345, 752: 346, 753: 347, 754: 348, 755: 349, 756: 350,
	757: 351, 758: 352, 759: 353, 760: 354, 761: 355, 762: 356, 763: 357, 764: 358,
	765: 359, 766: 360, 767: 361, 768: 362, 769: 363, 770: 364, 773: 366, 776: 368,
	777: 369, 778: 370, 779: 371, 782: 373, 783: 374, 784: 375, 785: 376, 786: 377,
	787: 378, 790: 380, 791: 381, 794: 383, 795: 384, 798: 386, 799: 387, 852: 389,
	853: 390, 854: 391, 855: 392, 856: 393, 857: 394, 858: 395, 859: 396, 860: 397,
	861: 398, 862: 399, 863: 400, 866: 402, 867: 403, 868: 404, 869: 405, 870: 406,
	871: 407, 872: 408, 873: 409, 874: 410, 875: 411, 876: 412, 877: 413, 878: 414,
	879: 415, 880: 416, 881: 417, 884: 419, 885: 420, 886: 421, 887: 422, 888: 423,
	889: 424, 892: 426, 893: 427, 894: 428, 895: 429, 896: 430, 897: 431, 898: 432,
	899: 433, 900: 434, 901: 435, 902: 436, 903: 437, 904: 438, 907: 440, 908: 441,
	911: 443, 912: 444, 913: 445, 914: 446, 915: 447, 916: 448, 919: 451, 920: 452,
	921: 453, 922: 454, 923: 455, 924: 456, 925: 457, 926: 458, 927: 459, 928: 460,
	929: 461, 930: 462, 933: 466, 934: 467, 935: 468, 936: 469, 937: 470, 938: 471,
	939: 472, 940: 473, 941: 474, 942: 475, 943: 476, 946: 478, 947: 479, 950: 481,
	951: 482, 952: 483, 953: 484, 954: 485, 955: 486, 956: 487, 957: 488, 958: 489,
	959: 490, 960: 491, 963: 493, 966: 495, 967: 496, 968: 497, 969: 498, 970: 499,
	971: 500, 972: 501, 973: 502, 974: 503, 975: 504, 976: 505, 977: 506, 978: 507,
	979: 508, 980: 509, 981: 510, 982: 511, 983: 512, 984: 513, 985: 514, 986: 515,
	989: 517, 990: 518, 991: 519, 992: 520, 993: 521, 994: 522, 997: 524, 998: 525,
	999: 526, 1000: 527, 1001: 528, 1004: 530, 1005: 531, 1006: 532, 1007: 533, 1008: 534,
	1009: 535, 1010: 536, 1011: 537, 1014: 539, 1015: 540, 1016: 541, 1017: 542, 1018: 543,
	1021: 567, 1024: 571, 1031: 611, 1032: 612, 1033: 613, 1034: 614, 1035: 615, 1036: 616,
	1037: 617, 1038: 618, 1039: 619, 1040: 620, 1041: 621, 1042: 622, 1043: 623, 1044: 624,
	1045: 625, 1046: 626, 1047: 627, 1048: 628, 1049: 629, 1050: 630, 1051: 631, 1052: 632,
	1053: 633, 1054: 634, 1055: 635, 1056: 636, 1057: 637, 1058: 638, 1059: 639, 1060: 640,
	1061: 641, 1062: 642, 1063: 643, 1064: 644, 1065: 645, 1066: 646, 1067: 647, 1068: 648,
	1069: 649, 1070: 650, 1071: 651, 1072: 652, 1073: 653, 1074: 654, 1075: 655, 1076: 656,
	1077: 657, 1078: 658, 1079: 659, 1080: 660, 1083: 662, 1084: 663, 1085: 664, 1086: 665,
	1087: 666, 1088: 667, 1089: 668, 1090: 669, 1091: 670, 1092: 671, 1093: 672, 1094: 673,
	1095: 674, 1096: 675, 1097: 676, 1098: 677, 1099: 678, 1100: 679, 1101: 680, 1102: 681,
	1103: 682, 1104: 683, 1105: 684, 1106: 685, 1107: 686, 1108: 687, 1109: 688, 1110: 689,
	1111: 690, 1114: 698, 1115: 699, 1118: 713, 1119: 714, 1120: 715, 1121: 716, 1122: 717,
	1123: 718, 1126: 721, 1127: 722, 1128: 723, 1129: 724, 1130: 725, 1131: 726, 1132: 727,
	1133: 728, 1134: 729, 1135: 730, 1136: 731, 1137: 732, 1138: 733, 1139: 734, 1140: 735,
	1141: 736, 1142: 737, 1143: 738, 1146: 744, 1150: 745, 1154: 748, 1177: 741, 1178: 743,
	1179: 742, 1183: 388, 1184: 754, 1185: 755, 1186: 756, 1187: 757, 1188: 758, 1189: 759,
	1190: 760, 1199: 761, 1213: 308, 1214: 544, 1215: 545, 1216: 546, 1217: 547, 1218: 548,
	1219: 549, 1220: 550, 1221: 551, 1222: 552, 1223: 553, 1224: 554, 1225: 555, 1226: 556,
	1227: 557, 1228: 558, 1229: 559, 1230: 560, 1231: 561, 1232: 562, 1233: 563, 1234: 564,
	1235: 565, 1236: 566, 1237: 568, 1238: 569, 1239: 570,
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

	// Read nickname (bytes 8-17)
	nickname := decodeGen3String(data[8:18])

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

	growthPos := typeToPos[0] * 12   // G = 0
	attacksPos := typeToPos[1] * 12  // A = 1

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

	// Get moves from Attacks substructure (bytes 0-7, 4 moves of 2 bytes each)
	moveNums := make([]int, 0, 4)
	for i := 0; i < 4; i++ {
		moveID := int(binary.LittleEndian.Uint16(decryptedData[attacksPos+i*2 : attacksPos+i*2+2]))
		if moveID > 0 {
			moveNums = append(moveNums, moveID)
		}
	}

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
		Nickname:   nickname,
		Level:      level,
		SpeciesNum: speciesNum,
		Nature:     nature,
		ItemNum:    itemNum,
		MoveNums:   moveNums,
		Stats:      stats,
		CurrentHP:  currentHP,
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
