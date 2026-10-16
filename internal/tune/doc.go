// Package tune holds the temporary tuning client: it reads the BHL
// sample and labels (bhl-abbr.sqlite), sends the sampled groups to
// abbr_names with features, caches the answers, re-scores them offline
// to fit gnames score weights, and calibrates the threshold.
//
// It is not part of the library and goes away when tuning is done.
package tune
