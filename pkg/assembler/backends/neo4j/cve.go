//
// Copyright 2023 The GUAC Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package neo4j

import (
	"context"
	"strings"

	"github.com/guacsec/guac/pkg/assembler/graphql/model"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

func (c *neo4jClient) Cve(ctx context.Context, cveSpec *model.CVESpec) ([]*model.Cve, error) {
	// fields: [year cveId cveId.id]
	fields := getPreloads(ctx)
	cveIDImplRequired := false
	for _, f := range fields {
		if f == cvdID {
			cveIDImplRequired = true
			break
		}
	}

	if !cveIDImplRequired {
		return c.cveYear(ctx, cveSpec)
	}

	session := c.driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close()

	var sb strings.Builder
	var firstMatch bool = true
	queryValues := map[string]any{}

	sb.WriteString("MATCH (root:Cve)-[:CveIsYear]->(cveYear:CveYear)-[:CveHasID]->(cveID:CveID)")

	setCveMatchValues(&sb, cveSpec, &firstMatch, queryValues)

	sb.WriteString(" RETURN cveYear.year, cveID.id")

	result, err := session.ReadTransaction(
		func(tx neo4j.Transaction) (interface{}, error) {
			// FIXME update to GHSA without root node.
			// result, err := tx.Run(sb.String(), queryValues)
			// if err != nil {
			// 	return nil, err
			// }

			// cvesPerYear := map[int][]*model.CVEId{}
			// for result.Next() {
			// 	cveID := &model.CVEId{
			// 		CveID: result.Record().Values[1].(string),
			// 	}
			// 	cvesPerYear[result.Record().Values[0].(int)] = append(cvesPerYear[result.Record().Values[0].(int)], cveID)
			// }
			// if err = result.Err(); err != nil {
			// 	return nil, err
			// }

			// cves := []*model.Cve{}
			// for year := range cvesPerYear {
			cve := &model.Cve{
				// Year:   year,
				// CveIds: cvesPerYear[year],
			}
			// cves = append(cves, cve)
			// }

			return cve, nil
		})
	if err != nil {
		return nil, err
	}

	return result.([]*model.Cve), nil
}

func (c *neo4jClient) cveYear(ctx context.Context, cveSpec *model.CVESpec) ([]*model.Cve, error) {
	session := c.driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close()

	var sb strings.Builder
	queryValues := map[string]any{}

	sb.WriteString("MATCH (n:Cve)-[:CveIsYear]->(cveYear:CveYear)")

	if cveSpec.Year != nil {
		matchProperties(&sb, true, "cveYear", "year", "$cveYear")
		queryValues["cveYear"] = cveSpec.Year
	}

	sb.WriteString(" RETURN cveYear.year")

	result, err := session.ReadTransaction(
		func(tx neo4j.Transaction) (interface{}, error) {
			// FIXME update to GHSA without root node.
			// result, err := tx.Run(sb.String(), queryValues)
			// if err != nil {
			// 	return nil, err
			// }

			// cves := []*model.Cve{}
			// for result.Next() {
			cve := &model.Cve{
				// Year: result.Record().Values[0].(int),
				// CveIds: []*model.CVEId{},
			}
			// 	cves = append(cves, cve)
			// }
			// if err = result.Err(); err != nil {
			// 	return nil, err
			// }

			return cve, nil
		})
	if err != nil {
		return nil, err
	}

	return result.([]*model.Cve), nil
}

func setCveMatchValues(sb *strings.Builder, cve *model.CVESpec, firstMatch *bool, queryValues map[string]any) {
	if cve != nil {
		if cve.Year != nil {
			matchProperties(sb, *firstMatch, "cveYear", "year", "$cveYear")
			queryValues["cveYear"] = cve.Year
			*firstMatch = false
		}

		if cve.CveID != nil {
			matchProperties(sb, *firstMatch, "cveID", "id", "$cveID")
			queryValues["cveID"] = strings.ToLower(*cve.CveID)
			*firstMatch = false
		}
	}
}

func (c *neo4jClient) IngestCve(ctx context.Context, cve *model.CVEInputSpec) (*model.Cve, error) {
	session := c.driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close()

	values := map[string]any{}
	values["year"] = cve.Year
	values["id"] = strings.ToLower(cve.CveID)

	result, err := session.WriteTransaction(
		func(tx neo4j.Transaction) (interface{}, error) {
			query := `MERGE (root:Cve)
MERGE (root) -[:CveIsYear]-> (cveYear:CveYear{year:$year})
MERGE (cveYear) -[:CveHasID]-> (cveID:CveID{id:$id})
RETURN cveYear.year, cveID.id`
			result, err := tx.Run(query, values)
			if err != nil {
				return nil, err
			}

			// query returns a single record
			record, err := result.Single()
			if err != nil {
				return nil, err
			}

			cveIdStr := record.Values[1].(string)
			yearStr := record.Values[0].(int)
			cve := generateModelCve(yearStr, cveIdStr)

			return cve, nil
		})
	if err != nil {
		return nil, err
	}

	return result.(*model.Cve), nil
}

// TODO: update to pass in the ID from neo4j
func generateModelCve(yearStr int, idStr string) *model.Cve {
	// FIXME update to GHSA without root node.
	// id := &model.CVEId{CveID: idStr}
	cve := model.Cve{
		Year: yearStr,
		// CveIds: []*model.CVEId{id},
	}
	return &cve
}
