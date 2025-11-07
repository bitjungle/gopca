# pollofpolls data

Average of national opinion polls on general elections in Norway.

Data is sourced from [pollofpolls.no](https://www.pollofpolls.no/?cmd=Stortinget&do=visallesnitt).

```sh 
curl -o pollofpolls_raw.csv "https://www.pollofpolls.no/lastned.csv?tabell=gallupsnitttabell&antall=0&type=riks&int=m&kommuneid=0&start=2000-01-01&slutt=$(date +%Y-%m-%d)"
```
This command downloads the latest data as a CSV file. Windows 1252 encoding. 

