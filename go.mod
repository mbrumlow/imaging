module github.com/mbrumlow/imaging

go 1.21

require github.com/mbrumlow/ppm v0.0.0-00010101000000-000000000000

// A small Netpbm decoder is bundled in-tree so the example programs build
// without network access. Point this at the upstream module if you prefer it.
replace github.com/mbrumlow/ppm => ./third_party/ppm
