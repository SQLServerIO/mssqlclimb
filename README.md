A Microsoft SQL Server utility to export data into different data formats with
support for templates.

This is a fork from [pgclimb](https://github.com/lukasmartinelli/pgclimb) 

Features:
- Export data to [JSON](#json-document), [JSON Lines](#json-lines), [CSV](#csv-and-tsv), [XLSX](#xlsx), [XML](#xml)
- Use [Templates](#templates) to support custom formats (HTML, Markdown, Text)

Use Cases:
- `SSIS` alternative for getting data out of SQL Server
- Publish data sets
- Create Excel reports from the database
- Generate HTML reports
- Export XML data for further processing with XSLT
- Transform data to JSON for graphing it with JavaScript libraries
- Generate readonly JSON APIs

## Install

**Install from source**

```bash
go get github.com/sqlserverio/mssqlclimb
```

## Supported Formats


### CSV and TSV

Exporting CSV and TSV files is very similar to using `psql` and the `COPY TO` statement.

```cmd
# Write CSV file to stdout with comma as default delimiter
mssqlclimb -c "SELECT * FROM employee_salaries" csv

# Save CSV file with custom delimiter and header row to file
mssqlclimb -o salaries.csv \
    -c "SELECT full_name, position_title FROM employee_salaries" \
    csv --delimiter ";" --header

# Create TSV file with SQL query from stdin
mssqlclimb -o positions.tsv tsv <<EOF
SELECT position_title, COUNT(*) FROM employee_salaries
GROUP BY position_title
ORDER BY 1
EOF
```

### JSON Document

Creating a single JSON document of a query is helpful if you
interface with other programs like providing data for JavaScript or creating
a readonly JSON API. You don't need to `json_agg` your objects, `mssqlclimb` will
automatically serialize the JSON for you - it also supports nested JSON objects for more complicated queries.

```bash
# Query all salaries into JSON array
mssqlclimb -c "SELECT * FROM employee_salaries" json

# Query all employees of a position as nested JSON object
cat << EOF > employees_by_position.sql
SELECT s.position_title, json_agg(s) AS employees
FROM employee_salaries s
GROUP BY s.position_title
ORDER BY 1
EOF

# Load query from file and store it as JSON array in file
mssqlclimb -f employees_by_position.sql \
    -o employees_by_position.json \
    json
```

### JSON Lines

[Newline delimited JSON](http://jsonlines.org/) is a good format to exchange
structured data in large quantities which does not fit well into the CSV format.
Instead of storing the entire JSON array each line is a valid JSON object.

```bash
# Query all salaries as separate JSON objects
mssqlclimb -c "SELECT * FROM employee_salaries" jsonlines

# In this example we interface with jq to pluck the first employee of each position
mssqlclimb -f employees_by_position.sql jsonlines | jq '.employees[0].full_name'
```

### XLSX

Excel files are really useful to exchange data with non programmers
and create graphs and filters. You can fill different datasets into different spreedsheets and distribute one single Excel file.

```bash
# Store all salaries in XLSX file
mssqlclimb -o salaries.xlsx -c "SELECT * FROM employee_salaries" xlsx

# Create XLSX file with multiple sheets
mssqlclimb -o salary_report.xlsx \
    -c "SELECT DISTINCT position_title FROM employee_salaries" \
    xlsx --sheet "positions"
mssqlclimb -o salary_report.xlsx \
    -c "SELECT full_name FROM employee_salaries" \
    xlsx --sheet "employees"
```

### XML

You can output XML to process it with other programs like [XLST](http://www.w3schools.com/xsl/).
To have more control over the XML output you should use the `mssqlclimb` template functionality directly to generate XML or build your own XML document with [XML functions in PostgreSQL](https://wiki.postgresql.org/wiki/XML_Support).

```bash
# Output XML for each row
mssqlclimb -o salaries.xml -c "SELECT * FROM employee_salaries" xml
```

A good default XML export is currently lacking because the XML format
can be controlled using templates.
If there is enough demand I will implement a solid
default XML support without relying on templates.

## Templates

Templates are the most powerful feature of `mssqlclimb` and allow you to implement
other formats that are not built in. In this example we will create a
HTML report of the salaries.

Create a template `salaries.tpl`.

```html
<!DOCTYPE html>
<html>
    <head><title>Montgomery County MD Employees</title></head>
    <body>
        <h2>Employees</h2>
        <ul>
            {{range .}}
            <li>{{.full_name}}</li>
            {{end}}
        </ul>
    </body>
</html>
```

And now run the template.

```
mssqlclimb -o salaries.html \
    -c "SELECT * FROM employee_salaries" \
    template salaries.tpl
```

## Database Connection

Database connection details can be provided via environment variables
or as separate flags (same flags as `psql`).

name        | default     | flags               | description
------------|-------------|---------------------|-----------------
`DB_NAME`   | `master`  | `-d`, `--dbname`    | database name
`DB_HOST`   | `localhost` | `--host`            | host name
`DB_PORT`   | `1433`      | `-p`, `--port`      | port
`DB_USER`   | `sa`        | `-U`, `--username`  | database user
`DB_PASS`   |             | `--pass`            | password (or empty if none)

## Advanced Use Cases

### Different ways of Querying

```bash
# Read query from stdin
echo "SELECT * FROM employee_salaries" | mssqlclimb
# Specify simple queries directly as arguments
mssqlclimb -c "SELECT * FROM employee_salaries"
# Load query from file
mssqlclimb -f query.sql
```

### Control Output

`mssqlclimb` will write the result to `stdout` by default.
By specifying the `-o` option you can write the output to a file.

```bash
mssqlclimb -o salaries.tsv -c "SELECT * FROM employee_salaries" tsv
```

# Contribute

## Dependencies

Go get the required dependencies for building `mssqlclimb`.

```bash
go get github.com/userve/cli
go get github.com/lib/go-mssqldb
go get github.com/jmoiron/sqlx
go get github.com/tealeg/xlsx
go get github.com/andrew-d/go-termutil
```

## Cross-compiling

We use [gox](https://github.com/mitchellh/gox) to create distributable
binaries for Windows, OSX and Linux.

```bash
docker run --rm -v "$(pwd)":/usr/src/mssqlclimb -w /usr/src/mssqlclimb tcnksm/gox:1.4.2-light
```

## Integration Tests

Run `test.sh` to run integration tests of the program with a SQL Server server. Take a look at the `.travis.yml`.
