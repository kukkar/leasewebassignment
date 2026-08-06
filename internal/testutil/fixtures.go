// Package testutil provides a small, deterministic server dataset shared by
// store- and API-level tests. It exists so integration tests assert against
// hand-computed expected values instead of re-deriving "expected" results
// with the same parsing code under test (which can hide regressions).
package testutil

import "github.com/sahil/leasewebassignment/internal/model"

// SampleServersCSV is a minimal CSV fixture (header + 10 rows) that exercises:
//   - multiple RAM values (4GB, 16GB, 32GB, 64GB, 128GB), each with the memory
//     technology suffix glued on with no separator exactly as the real
//     dataset does (e.g. "16GBDDR3"), so filter tests exercise the same
//     family-extraction normalization the real data requires
//   - all three disk-type families found in the real dataset (SAS, SSD, and
//     SATA2, which must normalize to the documented "SATA" filter value)
//   - multiple locations, including a repeated one, for substring matching
//   - a spread of total storage sizes for range-filter boundary testing
const SampleServersCSV = `Model,RAM,HDD,Location,Price
Dell R210,16GBDDR3,2x2TBSATA2,AmsterdamAMS-01,49.99
HP DL180,32GBDDR3,8x2TBSATA2,AmsterdamAMS-01,119.00
RH2288,128GBDDR4,4x480GBSSD,AmsterdamAMS-01,227.99
Dell R730XD,64GBDDR4,4x2TBSATA2,Washington D.C.WDC-01,319.99
HP DL180G6,32GBDDR3,8x300GBSAS,DallasDAL-10,170.99
Dell R930,64GBDDR4,2x120GBSSD,SingaporeSIN-11,1328.99
HP DL120G7,4GBDDR3,4x1TBSATA2,AmsterdamAMS-01,39.99
Dell R210-II,16GBDDR3,2x500GBSATA2,FrankfurtFRA-10,74.00
HP DL380eG8,64GBDDR3,8x2TBSATA2,FrankfurtFRA-10,165.99
Supermicro SC846,32GBDDR3,24x1TBSATA2,Washington D.C.WDC-01,421.99
`

// SampleServers is the parsed, ground-truth representation of SampleServersCSV.
// Total storage (GB), RAM family, and normalized disk type are documented per
// row so test expectations can be reasoned about without running the parser.
var SampleServers = []model.Server{
	{Model: "Dell R210", RAM: "16GBDDR3", HDD: "2x2TBSATA2", Location: "AmsterdamAMS-01", Price: 49.99},                // 16GB, 4096GB SATA
	{Model: "HP DL180", RAM: "32GBDDR3", HDD: "8x2TBSATA2", Location: "AmsterdamAMS-01", Price: 119.00},                // 32GB, 16384GB SATA
	{Model: "RH2288", RAM: "128GBDDR4", HDD: "4x480GBSSD", Location: "AmsterdamAMS-01", Price: 227.99},                 // 128GB, 1920GB SSD
	{Model: "Dell R730XD", RAM: "64GBDDR4", HDD: "4x2TBSATA2", Location: "Washington D.C.WDC-01", Price: 319.99},       // 64GB, 8192GB SATA
	{Model: "HP DL180G6", RAM: "32GBDDR3", HDD: "8x300GBSAS", Location: "DallasDAL-10", Price: 170.99},                 // 32GB, 2400GB SAS
	{Model: "Dell R930", RAM: "64GBDDR4", HDD: "2x120GBSSD", Location: "SingaporeSIN-11", Price: 1328.99},              // 64GB, 240GB SSD
	{Model: "HP DL120G7", RAM: "4GBDDR3", HDD: "4x1TBSATA2", Location: "AmsterdamAMS-01", Price: 39.99},                // 4GB, 4096GB SATA
	{Model: "Dell R210-II", RAM: "16GBDDR3", HDD: "2x500GBSATA2", Location: "FrankfurtFRA-10", Price: 74.00},           // 16GB, 1000GB SATA
	{Model: "HP DL380eG8", RAM: "64GBDDR3", HDD: "8x2TBSATA2", Location: "FrankfurtFRA-10", Price: 165.99},             // 64GB, 16384GB SATA
	{Model: "Supermicro SC846", RAM: "32GBDDR3", HDD: "24x1TBSATA2", Location: "Washington D.C.WDC-01", Price: 421.99}, // 32GB, 24576GB SATA
}
