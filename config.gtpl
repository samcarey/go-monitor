<html>
    <head>
    <title>FCC Frequency Tracker Configuration</title>
    <style>
        table, th, td {
            border: 1px solid black;
            border-collapse: collapse;
        }
        th, td {
            padding: 5px;
            text-align: left;
        }
    </style>
    </head>
    <body>
        <form action="/config" method="post">
            <table style="width:100%">
                <tr>
                    <th align="left">Parameter</th>
                    <th align="left">Value</th>
                </tr>
                <tr>
                    <td>Sample Rate (rows/sec)</td>
                    <td><input type="text" name="samp_rate"></td>
                </tr>
                <tr>
                    <td>Number of Rows</td>
                    <td><input type="text" name="nrows"></td>
                </tr>
                <tr>
                    <td>Database Name</td>
                    <td><input type="text" name="database_name"></td>
                </tr>
                <tr>
                    <td>Autoscale</td>
                    <td><input type="text" name="autoscale"></td>
                </tr>
                <tr>
                    <td>Min Val (without autoscale)</td>
                    <td><input type="text" name="min_val"></td>
                </tr>
                <tr>
                    <td>Max Val (without autoscale)</td>
                    <td><input type="text" name="max_val"></td>
                </tr>
                <tr>
                    <td>Number of Histogram Points</td>
                    <td><input type="text" name="nhist_points"></td>
                </tr>
            </table>
            <input type="submit" value="Go!">
        </form>
    </body>
</html>
