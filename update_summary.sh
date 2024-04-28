
echo "querying FY 2023-2024 stats"
DATASET=pvqr-7yc4
# curl 'https://data.cityofnewyork.us/resource/pvqr-7yc4.json?$select=issuing_agency,issue_date,count(*)+as+number_violations&$where=issuing_agency+in+('T','P','S')&$group=issuing_agency,issue_date&$having=number_violations+>+50&$limit=2000' > temp.json

# curl -s "https://data.cityofnewyork.us/resource/${DATASET}.json?\$select=issuing_agency,issue_date,count(*)+as+number_violations&\$where=issuing_agency+in+(%27T%27,%27P%27,%27S%27)&\$group=issuing_agency,issue_date&\$having=number_violations+%3E+50&\$limit=2000" > temp.json || echo "error" >&2

# https://data.cityofnewyork.us/resource/7mxj-7a6y.json?$select=issuing_agency,issue_date,count(*)+as+number_violations&$where=issuing_agency+in+('T','P','S','K')&$group=issuing_agency,issue_date&$having=number_violations+>+50&$limit=2000

jq -c '[.[] | select(.issuing_agency=="P")][-1]' temp.json
jq -c '[.[] | select(.issuing_agency=="T")][-1]' temp.json
jq -c '[.[] | select(.issuing_agency=="S")][-1]' temp.json

# `issue_date between '${dateFormat(startDate)}T00:00:00' and '${dateFormat(end)}T23:59:59'`

END_DATE=$(jq -r -c '[.[] | select(.issuing_agency=="P")][-1] | .issue_date' temp.json | awk -FT '{print $1}') # YYYY-mm-dd
START_DATE=$(date -j -v -11m -f %Y-%m-%d ${END_DATE} +%Y-%m-01)
DATASET_PREVIOUS="869v-vr48"

echo "fetching precinct stats from ${DATASET}, ${DATASET_PREVIOUS} - ${START_DATE} - ${END_DATE}"
set -o pipefail
function get_data() {
    curl -s "https://data.cityofnewyork.us/resource/${DATASET}.json?\$select=violation_precinct,count(*)+as+number_violations&\$where=issuing_agency+in+(%27P%27)+AND+issue_date+between+%27${START_DATE}T00:00:00%27+and+%27${END_DATE}T00:00:00%27&\$group=violation_precinct" | jq -c '.[]' || echo "error" >&2
    curl -s "https://data.cityofnewyork.us/resource/${DATASET_PREVIOUS}.json?\$select=violation_precinct,count(*)+as+number_violations&\$where=issuing_agency+in+(%27P%27)+AND+issue_date+between+%27${START_DATE}T00:00:00%27+and+%27${END_DATE}T00:00:00%27&\$group=violation_precinct" | jq -c '.[]' || echo "error" >&2
}

get_data | jq -c '{ (.violation_precinct ) : (.number_violations | tonumber)}' |  jq -c -n 'reduce (inputs | to_entries[]) as {$key,$value} ({}; .[$key] += $value) | to_entries[] | {violation_precinct:(.key | tonumber), number_violations:.value}' | jq -s > precinct_data.json