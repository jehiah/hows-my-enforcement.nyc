
<https://hows-my-enforcement.nyc/>

Is NYPD providing consistent and complete parking enforcement? Use this tool to audit the parking violations written on any block of NYC.


Data based on NYC Open Data for Parking Violations issued in each Fiscal Year.

https://data.cityofnewyork.us/City-Government/Parking-Violations-Issued-Fiscal-Year-2024/pvqr-7yc4


## Appendix


### Precinct -> Borough Mapping

```
curl -s https://www.nyc.gov/assets/nypd/data/precinct-house.json | jq -c '.features[] | {Precinct:.properties.PRECINCT, Borough:.properties.BORO}'
```