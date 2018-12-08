/* pgxsql

package = "integrationtest"

name = "DeletePerson"

resultType = "commandTag"

[[parameter]]
sql_name = "id"
allow_nil = true

pgxsql */

delete from person where id=:id;
