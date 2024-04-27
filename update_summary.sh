
echo "querying FY 2023-2024 stats"
DATSET=pvqr-7yc4
# curl 'https://data.cityofnewyork.us/resource/pvqr-7yc4.json?$select=issuing_agency,issue_date,count(*)+as+number_violations&$where=issuing_agency+in+('T','P','S')&$group=issuing_agency,issue_date&$having=number_violations+>+50&$limit=2000' > temp.json

curl -s "https://data.cityofnewyork.us/resource/${DATSET}.json?\$select=issuing_agency,issue_date,count(*)+as+number_violations&\$where=issuing_agency+in+(%27T%27,%27P%27,%27S%27)&\$group=issuing_agency,issue_date&\$having=number_violations+%3E+50&\$limit=2000" > temp.json || echo "error" >&2

# https://data.cityofnewyork.us/resource/7mxj-7a6y.json?$select=issuing_agency,issue_date,count(*)+as+number_violations&$where=issuing_agency+in+('T','P','S','K')&$group=issuing_agency,issue_date&$having=number_violations+>+50&$limit=2000

jq -c '[.[] | select(.issuing_agency=="P")][-1]' temp.json
jq -c '[.[] | select(.issuing_agency=="T")][-1]' temp.json
jq -c '[.[] | select(.issuing_agency=="S")][-1]' temp.json

# `issue_date between '${dateFormat(startDate)}T00:00:00' and '${dateFormat(end)}T23:59:59'`

# DATE=$(jq -r -c '[.[] | select(.issuing_agency=="P")][-1] | .issue_date' temp.json)

curl -s "https://data.cityofnewyork.us/resource/${DATSET}.json?\$select=violation_precinct,count(*)+as+number_violations&\$where=issuing_agency+in+(%27P%27)&\$group=violation_precinct" > precinct_data.json || echo "error" >&2
