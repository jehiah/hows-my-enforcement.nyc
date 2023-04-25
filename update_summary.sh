
# FY 2022-2023
echo "querying FY 2022-2023 stats"
# curl 'https://data.cityofnewyork.us/resource/pvqr-7yc4.json?$select=issuing_agency,issue_date,count(*)+as+number_violations&$where=issuing_agency+in+('T','P','S')&$group=issuing_agency,issue_date&$having=number_violations+>+50&$limit=2000' > temp.json

curl -s 'https://data.cityofnewyork.us/resource/pvqr-7yc4.json?$select=issuing_agency,issue_date,count(*)+as+number_violations&$where=issuing_agency+in+(%27T%27,%27P%27,%27S%27)&$group=issuing_agency,issue_date&$having=number_violations+%3E+50&$limit=2000' > temp.json || echo "error" >&2

# https://data.cityofnewyork.us/resource/7mxj-7a6y.json?$select=issuing_agency,issue_date,count(*)+as+number_violations&$where=issuing_agency+in+('T','P','S','K')&$group=issuing_agency,issue_date&$having=number_violations+>+50&$limit=2000

jq -c '[.[] | select(.issuing_agency=="P")][-1]' temp.json
jq -c '[.[] | select(.issuing_agency=="T")][-1]' temp.json
jq -c '[.[] | select(.issuing_agency=="S")][-1]' temp.json
