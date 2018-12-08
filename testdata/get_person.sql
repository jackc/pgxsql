/* pgxsql
package = "main"

name = "GetPerson"

resultType = "row"

[[parameter]]
sql_name = "id"
allow_nil = true

[[column]]
sql_name = "id"
go_public_name = "ID"
go_private_name = "id"

[[column]]
sql_name = "name"
go_public_name = "Name"
go_private_name = "name"

pgxsql */

select id, name
from person
where id=:id;
