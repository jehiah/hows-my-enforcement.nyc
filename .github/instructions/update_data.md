Monthly Data Update

Reports query open data for fixed timeranges - the open data dataset updates on the 16th of the month with variying data availability by agency.

Note: the dataset `FY2024` is updated in perpetuity now.

# Query open data for date ranges

1. run `./update_summary.sh` This will output the dates agencies have data through. This will also modify `precinct_data.json` to be committed
1. update end dates for `FY2024` in `report.html`
1. update end dates for `FY2024` in `precinct.html`

# Update index

1. run `cd update_data; go build && ./update_data`. This will output a new `DatasetID`
1. update `index.html` with new `DatasetID`
