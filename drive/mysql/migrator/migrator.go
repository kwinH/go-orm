package migrator

import (
	"fmt"
	"github.com/kwinh/go-orm/schema"
	"sort"
	"strings"
)

// Migrator m struct
type Migrator struct {
	DB schema.IDBParse
}

func (m Migrator) TableExist(tableName string) bool {
	sql := fmt.Sprintf("SHOW TABLES LIKE '%v'", tableName)
	res, _ := m.DB.Query(sql)
	defer res.Close()

	var table string
	res.Next()
	_ = res.Scan(&table)

	if table != "" {
		return true
	}
	return false
}

func (m Migrator) TableInfo(tableName string) schema.TableInfo {
	sql := fmt.Sprintf("SHOW CREATE TABLE `%v`", tableName)
	res, _ := m.DB.Query(sql)
	defer res.Close()
	res.Next()

	_ = res.Scan(&tableName, &sql)

	primaryKey := ""
	fieldsInfo := make(map[string]string)
	uniqueKeys := make(map[string][]string)
	indexKeys := make(map[string][]string)
	fullKeys := make(map[string][]string)

	collate := ""
	for _, fieldInfo := range strings.Split(sql[strings.Index(sql, "\n")+1:], "\n") {
		if strings.Index(fieldInfo, " `") == 1 {
			fieldNameIndex := strings.Index(fieldInfo, "` ")
			fieldName := fieldInfo[3:fieldNameIndex]
			if collate == "" {
				collateIndex := strings.Index(fieldInfo, "COLLATE ")
				if collateIndex != -1 {
					collate = fieldInfo[collateIndex+8:]
					collate = " COLLATE " + collate[:strings.Index(collate, " ")]
				}
			}

			fieldsInfo[fieldName] = fieldInfo[2 : len(fieldInfo)-1]

			if collate != "" {
				fieldsInfo[fieldName] = strings.ReplaceAll(fieldsInfo[fieldName], collate, "")
			}
		} else {
			var indexType schema.IndexType
			indexKey := ""
			if strings.Index(fieldInfo, " PRIMARY KEY ") == 1 {
				indexType = schema.PrimaryKey
				fieldInfo = fieldInfo[15:]
			} else if strings.Index(fieldInfo, " UNIQUE KEY ") == 1 {
				indexType = schema.UNIQUEKEY
				fieldInfo = fieldInfo[14:]
				i := strings.Index(fieldInfo, "` ")
				indexKey = fieldInfo[:i]
				fieldInfo = fieldInfo[i+3:]
			} else if strings.Index(fieldInfo, " FULLTEXT KEY ") == 1 {
				indexType = schema.FULLTEXTKEY
				fieldInfo = fieldInfo[16:]
				i := strings.Index(fieldInfo, "` ")
				indexKey = fieldInfo[:i]
				fieldInfo = fieldInfo[i+3:]
			} else if strings.Index(fieldInfo, " KEY ") == 1 {
				indexType = schema.INDEXKEY
				fieldInfo = fieldInfo[7:]
				i := strings.Index(fieldInfo, "` ")
				indexKey = fieldInfo[:i]
				fieldInfo = fieldInfo[i+3:]
			}

			if indexType != "" {
				fieldInfo = fieldInfo[:strings.Index(fieldInfo, ")")]
				fieldInfo = strings.ReplaceAll(fieldInfo, "`", "")

				switch indexType {
				case schema.UNIQUEKEY:
					uniqueKeys[indexKey] = strings.Split(fieldInfo, ",")
				case schema.INDEXKEY:
					indexKeys[indexKey] = strings.Split(fieldInfo, ",")
				case schema.FULLTEXTKEY:
					fullKeys[indexKey] = strings.Split(fieldInfo, ",")
				case schema.PrimaryKey:
					primaryKey = strings.Split(fieldInfo, ",")[0]
				}
			}
		}
	}

	return schema.TableInfo{
		FieldsInfo: fieldsInfo,
		PrimaryKey: primaryKey,
		UniqueKeys: uniqueKeys,
		IndexKeys:  indexKeys,
		FullKeys:   fullKeys,
	}
}

func (m Migrator) getFieldSql(field *schema.Field) string {
	var (
		notNull       string
		defaultValue  string
		autoIncrement string
	)

	if field.AutoIncrement {
		autoIncrement = " AUTO_INCREMENT"
		notNull = " NOT NULL"
	}

	if field.DataType == schema.String {
		if field.DefaultValue != schema.DefaultNull {
			field.DefaultValue = "''"
		}
	}

	if field.HavDefaultValue == true {
		defaultValue = fmt.Sprintf(" DEFAULT %v", field.DefaultValue)
	} else if field.DataType == schema.Time ||
		field.DataType == schema.Json ||
		(field.DataType == schema.String && field.Size >= 65536) {
		field.DefaultValue = schema.DefaultNull
		defaultValue = fmt.Sprintf(" DEFAULT NULL")
	}

	if field.DefaultValue != schema.DefaultNull {
		notNull = " NOT NULL"
	}

	if field.Comment != "" {
		field.Comment = fmt.Sprintf(" COMMENT '%s'", field.Comment)
	}

	sql := fmt.Sprintf("`%s` %s%s%s%s%s", field.FieldName, field.Type, notNull, autoIncrement, defaultValue, field.Comment)
	return sql
}

func (m Migrator) getIndex(indexType schema.IndexType, indexFields schema.IndexList) string {
	indexSql := ""
	for key, fields := range indexFields {
		sort.Slice(fields, func(i, j int) bool { return fields[i].Priority < fields[j].Priority })
		fieldsLen := len(fields)

		sql := ""
		if indexType == schema.PrimaryKey {
			sql = fmt.Sprintf("%s (%s", indexType, strings.Repeat("`%s`,", fieldsLen))
		} else {
			sql = fmt.Sprintf("%s `%s` (%s", indexType, key, strings.Repeat("`%s`,", fieldsLen))
		}

		sql = strings.Trim(sql, ",") + "),\n"
		fieldNames := make([]any, fieldsLen)
		for i, field := range fields {
			fieldNames[i] = field.Field.FieldName
		}

		indexSql += fmt.Sprintf(sql, fieldNames...)
	}
	return indexSql
}

func (m Migrator) Create(schema1 *schema.Schema) error {
	sql := "CREATE TABLE `" + schema1.TableName + "` (\n"
	for _, field := range schema1.Fields {
		sql += m.getFieldSql(field) + ",\n"
	}

	if schema1.PrimaryKey != nil {
		sql += "PRIMARY KEY (`" + schema1.PrimaryKey.FieldName + "`),\n"
	}

	sql += m.getIndex(schema.UNIQUEKEY, schema1.UniqueKeys)
	sql += m.getIndex(schema.INDEXKEY, schema1.IndexKeys)
	sql += m.getIndex(schema.FULLTEXTKEY, schema1.FullKeys)

	sql = strings.Trim(sql, ",\n")
	sql += "\n)"

	_, err := m.DB.Exec(sql)

	if err != nil {
		return err
	}
	return nil
}

func (m Migrator) AddField(TableName string, field *schema.Field) error {
	sql := m.getFieldSql(field)
	sql = fmt.Sprintf("ALTER TABLE `%s` ADD %s", TableName, sql)
	_, err := m.DB.Exec(sql)

	if err != nil {
		return err
	}

	return nil
}

func (m Migrator) ModifyField(TableName string, field *schema.Field) error {
	sql := m.getFieldSql(field)
	sql = fmt.Sprintf("ALTER TABLE `%s` MODIFY %s", TableName, sql)
	_, err := m.DB.Exec(sql)

	if err != nil {
		return err
	}

	return nil
}

func (m Migrator) DropField(TableName string, FiledName string) error {
	sql := fmt.Sprintf("ALTER TABLE `%s` DROP `%s`", TableName, FiledName)
	_, err := m.DB.Exec(sql)

	if err != nil {
		return err
	}

	return nil
}

func (m Migrator) AddIndex(tableName string, indexType schema.IndexType, indexFields schema.IndexList) error {

	sql := fmt.Sprintf("ALTER TABLE `%s` ADD %s", tableName, strings.Trim(m.getIndex(indexType, indexFields), ",\n"))
	_, err := m.DB.Exec(sql)

	if err != nil {
		return err
	}

	return nil
}

func (m Migrator) DropIndex(indexKey, tableName string) error {

	sql := fmt.Sprintf("DROP INDEX `%s` ON `%s`", indexKey, tableName)
	_, err := m.DB.Exec(sql)

	if err != nil {
		return err
	}

	return nil
}

func (m Migrator) DropPrimaryIndex(tableName string) error {

	sql := fmt.Sprintf("alter table `%s` drop primary key", tableName)
	_, err := m.DB.Exec(sql)

	if err != nil {
		return err
	}

	return nil
}

func (m Migrator) UpdateIndex(schema1 *schema.Schema, schemaKeys schema.IndexList, keys map[string][]string, modify bool, indexType schema.IndexType) (err error) {
	for key, fields := range schemaKeys {
		sort.Slice(fields, func(i, j int) bool { return fields[i].Priority < fields[j].Priority })
		if fields1, ok := keys[key]; !ok {
			indexFields := make(schema.IndexList)
			indexFields[key] = fields

			err = m.AddIndex(schema1.TableName, indexType, indexFields)
			if err != nil {
				return err
			}
		} else {
			if modify {
				if len(fields1) == len(fields) {
					continue
				}

				for i, field := range fields1 {
					if field != fields[i].Field.FieldName {
						indexFields := make(schema.IndexList)
						indexFields[key] = fields
						err = m.DropIndex(key, schema1.TableName)
						if err != nil {
							return err
						}
						err = m.AddIndex(schema1.TableName, indexType, indexFields)
						if err != nil {
							return err
						}
						break
					}
				}
			}
		}
	}
	return
}

func (m Migrator) Auto(value any, modify, drop bool) error {
	schema1 := m.DB.Parse(value)
	if !m.TableExist(schema1.TableName) {
		return m.Create(schema1)
	}

	tableInfo := m.TableInfo(schema1.TableName)

	for _, field := range schema1.Fields {
		if fieldInfo, ok := tableInfo.FieldsInfo[field.FieldName]; !ok {
			err := m.AddField(schema1.TableName, field)
			if err != nil {
				return err
			}
		} else {
			if modify {
				if fieldInfo != m.getFieldSql(field) {
					err := m.ModifyField(schema1.TableName, field)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	if drop {
		for fieldName, _ := range tableInfo.FieldsInfo {
			if schema1.GetField(fieldName) == nil {
				err := m.DropField(schema1.TableName, fieldName)
				if err != nil {
					return err
				}
			}
		}

		err := m.DropIndexList(tableInfo.UniqueKeys, schema1.UniqueKeys, schema1.TableName)
		if err != nil {
			return err
		}

		err = m.DropIndexList(tableInfo.IndexKeys, schema1.IndexKeys, schema1.TableName)
		if err != nil {
			return err
		}

		err = m.DropIndexList(tableInfo.FullKeys, schema1.FullKeys, schema1.TableName)
		if err != nil {
			return err
		}

		if schema1.PrimaryKey == nil && tableInfo.PrimaryKey != "" {
			err := m.DropPrimaryIndex(schema1.TableName)
			if err != nil {
				return err
			}
		}
	}

	if modify {
		if tableInfo.PrimaryKey != schema1.PrimaryKey.FieldName {
			if tableInfo.PrimaryKey != "" {
				err := m.DropPrimaryIndex(schema1.TableName)
				if err != nil {
					return err
				}
			}

			indexFields := make(schema.IndexList)
			fields := make([]schema.Index, 1)
			fields[0] = schema.Index{
				Priority: 0,
				Field:    schema1.PrimaryKey,
			}
			indexFields["primaryKey"] = fields
			err := m.AddIndex(schema1.TableName, schema.PrimaryKey, indexFields)
			if err != nil {
				return err
			}
		}
	}

	err := m.UpdateIndex(schema1, schema1.UniqueKeys, tableInfo.UniqueKeys, modify, schema.UNIQUEKEY)
	if err != nil {
		return err
	}

	err = m.UpdateIndex(schema1, schema1.IndexKeys, tableInfo.IndexKeys, modify, schema.INDEXKEY)
	if err != nil {
		return err
	}

	err = m.UpdateIndex(schema1, schema1.FullKeys, tableInfo.IndexKeys, modify, schema.FULLTEXTKEY)
	if err != nil {
		return err
	}

	return nil
}

func (m Migrator) DropIndexList(keys map[string][]string, schemaKeys schema.IndexList, tableName string) error {
	for key, _ := range keys {
		if _, ok := schemaKeys[key]; !ok {
			err := m.DropIndex(key, tableName)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
