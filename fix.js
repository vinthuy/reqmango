const fs = require('fs');
const p = 'd:\\code\\reqmango\\backend\\internal\\rql\\executor.go';
let c = fs.readFileSync(p, 'utf8');
const target = '\tswitch mapping.FieldType {\n\tcase "state_group":';
const replacement = '\tswitch mapping.FieldType {\n\tcase "number":\n\t\tif mapping.JoinTable == "states" && mapping.JoinKey == "name" {\n\t\t\treturn &rawCondition{\n\t\t\t\tSQL:  fmt.Sprintf("%s %s (SELECT id FROM %s WHERE name IN (%s))", mapping.ColumnName, op, mapping.JoinTable, placeholderList),\n\t\t\t\tArgs: args,\n\t\t\t}, nil\n\t\t}\n\tcase "state_group":';
if (c.includes(target)) {
    c = c.replace(target, replacement);
    fs.writeFileSync(p, c, 'utf8');
    console.log('SUCCESS');
} else {
    console.log('TARGET NOT FOUND');
}
