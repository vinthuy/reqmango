$p = "d:\code\reqmango\backend\internal\rql\executor.go"
$c = [System.IO.File]::ReadAllText($p)
$target = 'switch mapping.FieldType {
case "state_group":'
$replacement = 'switch mapping.FieldType {
case "number":
if mapping.JoinTable == "states" && mapping.JoinKey == "name" {
return &rawCondition{
SQL:  fmt.Sprintf("%s %s (SELECT id FROM %s WHERE name IN (%s))", mapping.ColumnName, op, mapping.JoinTable, placeholderList),
Args: args,
}, nil
}
case "state_group":'
$c2 = $c.Replace($target, $replacement)
if ($c2 -ne $c) {
    [System.IO.File]::WriteAllText($p, $c2)
    Write-Host "SUCCESS: executor.go patched"
} else {
    Write-Host "ERROR: target not found"
}
