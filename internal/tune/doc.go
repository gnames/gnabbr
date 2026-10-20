// Package tune holds the temporary tuning client. It reads the BHL
// sample exported as gnabbr input (abbrctx/export.py: docs/<item>.json,
// items.jsonl, labels.jsonl), reconciles the documents with the gnabbr
// core through a cache of abbr_names answers, and compares expanded
// names with the labels.
//
// Next: re-score cached answers offline from features to fit gnames
// score weights, and calibrate the threshold.
//
// It is not part of the library and goes away when tuning is done.
package tune
