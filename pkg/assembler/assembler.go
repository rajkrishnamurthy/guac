//
// Copyright 2022 The GUAC Authors.
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

package assembler

import (
	"context"
	"strings"

	"github.com/guacsec/guac/pkg/assembler/clients/generated"
	"github.com/guacsec/guac/pkg/assembler/helpers"
)

type assembler struct{} //nolint: unused

// IngestPredicates contains the set of predicates that want to be
// ingested based on the GUAC ontology. It only has evidence trees as
// ingestion of the software trees are implicit and handled by the
// client library.
type IngestPredicates struct {
	CertifyScorecard []CertifyScorecardIngest
	IsDependency     []IsDependencyIngest
	IsOccurrence     []IsOccurrenceIngest
	HasSlsa          []HasSlsaIngest
	CertifyVuln      []CertifyVulnIngest
	IsVuln           []IsVulnIngest
	HasSourceAt      []HasSourceAtIngest
	CertifyBad       []CertifyBadIngest
	CertifyGood      []CertifyGoodIngest
	HasSBOM          []HasSBOMIngest
	HashEqual        []HashEqualIngest
	PkgEqual         []PkgEqualIngest
	Vex              []VexIngest
}

type CertifyScorecardIngest struct {
	Source    *generated.SourceInputSpec
	Scorecard *generated.ScorecardInputSpec
}

type IsDependencyIngest struct {
	Pkg          *generated.PkgInputSpec
	DepPkg       *generated.PkgInputSpec
	IsDependency *generated.IsDependencyInputSpec
}

type IsOccurrenceIngest struct {
	// Occurrence describes either pkg or src
	Pkg *generated.PkgInputSpec
	Src *generated.SourceInputSpec

	// Artifact is the required object of the occurence
	Artifact *generated.ArtifactInputSpec

	IsOccurrence *generated.IsOccurrenceInputSpec
}

type HasSlsaIngest struct {
	Artifact  *generated.ArtifactInputSpec
	HasSlsa   *generated.SLSAInputSpec
	Materials []generated.ArtifactInputSpec
	Builder   *generated.BuilderInputSpec

	// Upon more investigation, seems like SLSA should
	// only be applied to an artifact and linkages to pkg
	// or src should be done via IsOccurrence
	// Pkg      *generated.PkgInputSpec
	// Src      *generated.SourceInputSpec
}

type CertifyVulnIngest struct {
	// pkg is required
	Pkg *generated.PkgInputSpec

	// vulnerability should be either OSV, CVE, GHSA, or none if no vulnerability is found
	OSV  *generated.OSVInputSpec
	CVE  *generated.CVEInputSpec
	GHSA *generated.GHSAInputSpec

	VulnData *generated.VulnerabilityMetaDataInput
}

// Only CVE or GHSA needed, not both
type IsVulnIngest struct {
	OSV    *generated.OSVInputSpec
	CVE    *generated.CVEInputSpec
	GHSA   *generated.GHSAInputSpec
	IsVuln *generated.IsVulnerabilityInputSpec
}

type HasSourceAtIngest struct {
	Pkg          *generated.PkgInputSpec
	PkgMatchFlag generated.MatchFlags
	Src          *generated.SourceInputSpec
	HasSourceAt  *generated.HasSourceAtInputSpec
}

type CertifyBadIngest struct {
	// certifyBad describes either pkg, src or artifact
	Pkg          *generated.PkgInputSpec
	PkgMatchFlag generated.MatchFlags
	Src          *generated.SourceInputSpec
	Artifact     *generated.ArtifactInputSpec
	CertifyBad   *generated.CertifyBadInputSpec
}

type CertifyGoodIngest struct {
	// certifyGood describes either pkg, src or artifact
	Pkg          *generated.PkgInputSpec
	PkgMatchFlag generated.MatchFlags
	Src          *generated.SourceInputSpec
	Artifact     *generated.ArtifactInputSpec
	CertifyGood  *generated.CertifyGoodInputSpec
}

type HasSBOMIngest struct {
	// hasSBOM describes either pkg or artifact
	Pkg      *generated.PkgInputSpec
	Artifact *generated.ArtifactInputSpec

	HasSBOM *generated.HasSBOMInputSpec
}

type VexIngest struct {
	// pkg or artifact is required
	Pkg      *generated.PkgInputSpec
	Artifact *generated.ArtifactInputSpec

	// vulnerability should be either OSV, CVE, GHSA
	OSV  *generated.OSVInputSpec
	CVE  *generated.CVEInputSpec
	GHSA *generated.GHSAInputSpec

	VexData *generated.VexStatementInputSpec
}

type HashEqualIngest struct {
	// HashEqualIngest describes two artifacts are the same
	Artifact      *generated.ArtifactInputSpec
	EqualArtifact *generated.ArtifactInputSpec

	HashEqual *generated.HashEqualInputSpec
}

type PkgEqualIngest struct {
	// PkgEqualIngest describes two packages are the same
	Pkg      *generated.PkgInputSpec
	EqualPkg *generated.PkgInputSpec
	PkgEqual *generated.PkgEqualInputSpec
}

func (i IngestPredicates) GetPackages(ctx context.Context) []*generated.PkgInputSpec {
	packageMap := make(map[string]*generated.PkgInputSpec)
	for _, dep := range i.IsDependency {
		if dep.Pkg != nil {
			pkgPurl := helpers.PkgInputSpecToPurl(dep.Pkg)
			if _, ok := packageMap[pkgPurl]; !ok {
				packageMap[pkgPurl] = dep.Pkg
			}
		}
		if dep.DepPkg != nil {
			depPkgPurl := helpers.PkgInputSpecToPurl(dep.DepPkg)
			if _, ok := packageMap[depPkgPurl]; !ok {
				packageMap[depPkgPurl] = dep.DepPkg
			}
		}
	}
	for _, occur := range i.IsOccurrence {
		if occur.Pkg != nil {
			pkgPurl := helpers.PkgInputSpecToPurl(occur.Pkg)
			if _, ok := packageMap[pkgPurl]; !ok {
				packageMap[pkgPurl] = occur.Pkg
			}
		}
	}
	for _, vuln := range i.CertifyVuln {
		if vuln.Pkg != nil {
			pkgPurl := helpers.PkgInputSpecToPurl(vuln.Pkg)
			if _, ok := packageMap[pkgPurl]; !ok {
				packageMap[pkgPurl] = vuln.Pkg
			}
		}
	}
	for _, hasSource := range i.HasSourceAt {
		if hasSource.Pkg != nil {
			pkgPurl := helpers.PkgInputSpecToPurl(hasSource.Pkg)
			if _, ok := packageMap[pkgPurl]; !ok {
				packageMap[pkgPurl] = hasSource.Pkg
			}
		}
	}
	for _, bad := range i.CertifyBad {
		if bad.Pkg != nil {
			pkgPurl := helpers.PkgInputSpecToPurl(bad.Pkg)
			if _, ok := packageMap[pkgPurl]; !ok {
				packageMap[pkgPurl] = bad.Pkg
			}
		}
	}
	for _, good := range i.CertifyGood {
		if good.Pkg != nil {
			pkgPurl := helpers.PkgInputSpecToPurl(good.Pkg)
			if _, ok := packageMap[pkgPurl]; !ok {
				packageMap[pkgPurl] = good.Pkg
			}
		}
	}
	for _, sbom := range i.HasSBOM {
		if sbom.Pkg != nil {
			pkgPurl := helpers.PkgInputSpecToPurl(sbom.Pkg)
			if _, ok := packageMap[pkgPurl]; !ok {
				packageMap[pkgPurl] = sbom.Pkg
			}
		}
	}
	for _, v := range i.Vex {
		if v.Pkg != nil {
			pkgPurl := helpers.PkgInputSpecToPurl(v.Pkg)
			if _, ok := packageMap[pkgPurl]; !ok {
				packageMap[pkgPurl] = v.Pkg
			}
		}
	}
	for _, equal := range i.PkgEqual {
		if equal.Pkg != nil {
			pkgPurl := helpers.PkgInputSpecToPurl(equal.Pkg)
			if _, ok := packageMap[pkgPurl]; !ok {
				packageMap[pkgPurl] = equal.Pkg
			}
		}
		if equal.EqualPkg != nil {
			equalPkgPurl := helpers.PkgInputSpecToPurl(equal.EqualPkg)
			if _, ok := packageMap[equalPkgPurl]; !ok {
				packageMap[equalPkgPurl] = equal.EqualPkg
			}
		}
	}
	packages := make([]*generated.PkgInputSpec, 0, len(packageMap))

	for _, pkg := range packageMap {
		packages = append(packages, pkg)
	}
	return packages
}

func (i IngestPredicates) GetSources(ctx context.Context) []*generated.SourceInputSpec {
	sourceMap := make(map[string]*generated.SourceInputSpec)
	for _, score := range i.CertifyScorecard {
		if score.Source != nil {
			sourceString := concatenateSourceInput(score.Source)
			if _, ok := sourceMap[sourceString]; !ok {
				sourceMap[sourceString] = score.Source
			}
		}
	}
	for _, occur := range i.IsOccurrence {
		if occur.Src != nil {
			sourceString := concatenateSourceInput(occur.Src)
			if _, ok := sourceMap[sourceString]; !ok {
				sourceMap[sourceString] = occur.Src
			}
		}
	}
	for _, hasSource := range i.HasSourceAt {
		if hasSource.Src != nil {
			sourceString := concatenateSourceInput(hasSource.Src)
			if _, ok := sourceMap[sourceString]; !ok {
				sourceMap[sourceString] = hasSource.Src
			}
		}
	}
	for _, bad := range i.CertifyBad {
		if bad.Src != nil {
			sourceString := concatenateSourceInput(bad.Src)
			if _, ok := sourceMap[sourceString]; !ok {
				sourceMap[sourceString] = bad.Src
			}
		}
	}
	for _, good := range i.CertifyGood {
		if good.Src != nil {
			sourceString := concatenateSourceInput(good.Src)
			if _, ok := sourceMap[sourceString]; !ok {
				sourceMap[sourceString] = good.Src
			}
		}
	}
	sources := make([]*generated.SourceInputSpec, 0, len(sourceMap))

	for _, source := range sourceMap {
		sources = append(sources, source)
	}
	return sources
}

func (i IngestPredicates) GetArtifacts(ctx context.Context) []*generated.ArtifactInputSpec {
	artifactMap := make(map[string]*generated.ArtifactInputSpec)
	for _, occur := range i.IsOccurrence {
		if occur.Artifact != nil {
			artifactString := occur.Artifact.Algorithm + ":" + occur.Artifact.Digest
			if _, ok := artifactMap[artifactString]; !ok {
				artifactMap[artifactString] = occur.Artifact
			}
		}
	}
	for _, slsa := range i.HasSlsa {
		if slsa.Artifact != nil {
			artifactString := slsa.Artifact.Algorithm + ":" + slsa.Artifact.Digest
			if _, ok := artifactMap[artifactString]; !ok {
				artifactMap[artifactString] = slsa.Artifact
			}
		}
	}
	for _, sbom := range i.HasSBOM {
		if sbom.Artifact != nil {
			artifactString := sbom.Artifact.Algorithm + ":" + sbom.Artifact.Digest
			if _, ok := artifactMap[artifactString]; !ok {
				artifactMap[artifactString] = sbom.Artifact
			}
		}
	}
	for _, bad := range i.CertifyBad {
		if bad.Artifact != nil {
			artifactString := bad.Artifact.Algorithm + ":" + bad.Artifact.Digest
			if _, ok := artifactMap[artifactString]; !ok {
				artifactMap[artifactString] = bad.Artifact
			}
		}
	}
	for _, good := range i.CertifyGood {
		if good.Artifact != nil {
			artifactString := good.Artifact.Algorithm + ":" + good.Artifact.Digest
			if _, ok := artifactMap[artifactString]; !ok {
				artifactMap[artifactString] = good.Artifact
			}
		}
	}
	for _, v := range i.Vex {
		if v.Artifact != nil {
			artifactString := v.Artifact.Algorithm + ":" + v.Artifact.Digest
			if _, ok := artifactMap[artifactString]; !ok {
				artifactMap[artifactString] = v.Artifact
			}
		}
	}
	for _, equal := range i.HashEqual {
		if equal.Artifact != nil {
			artifactString := equal.Artifact.Algorithm + ":" + equal.Artifact.Digest
			if _, ok := artifactMap[artifactString]; !ok {
				artifactMap[artifactString] = equal.Artifact
			}
		}
		if equal.EqualArtifact != nil {
			artifactString := equal.EqualArtifact.Algorithm + ":" + equal.EqualArtifact.Digest
			if _, ok := artifactMap[artifactString]; !ok {
				artifactMap[artifactString] = equal.EqualArtifact
			}
		}
	}
	artifacts := make([]*generated.ArtifactInputSpec, 0, len(artifactMap))

	for _, art := range artifactMap {
		artifacts = append(artifacts, art)
	}
	return artifacts
}

func (i IngestPredicates) GetMaterials(ctx context.Context) []generated.ArtifactInputSpec {
	materialMap := make(map[string]generated.ArtifactInputSpec)
	for _, slsa := range i.HasSlsa {
		for _, mat := range slsa.Materials {
			artifactString := mat.Algorithm + ":" + mat.Digest
			if _, ok := materialMap[artifactString]; !ok {
				materialMap[artifactString] = mat
			}
		}

	}
	materials := make([]generated.ArtifactInputSpec, 0, len(materialMap))

	for _, mat := range materialMap {
		materials = append(materials, mat)
	}
	return materials
}

func (i IngestPredicates) GetBuilders(ctx context.Context) []*generated.BuilderInputSpec {
	builderMap := make(map[string]*generated.BuilderInputSpec)
	for _, slsa := range i.HasSlsa {
		if slsa.Builder != nil {
			if _, ok := builderMap[slsa.Builder.Uri]; !ok {
				builderMap[slsa.Builder.Uri] = slsa.Builder
			}
		}
	}
	builders := make([]*generated.BuilderInputSpec, 0, len(builderMap))

	for _, build := range builderMap {
		builders = append(builders, build)
	}
	return builders
}

func (i IngestPredicates) GetCVEs(ctx context.Context) []*generated.CVEInputSpec {
	cveMap := make(map[string]*generated.CVEInputSpec)
	for _, vuln := range i.CertifyVuln {
		if vuln.CVE != nil {
			if _, ok := cveMap[vuln.CVE.CveId]; !ok {
				cveMap[vuln.CVE.CveId] = vuln.CVE
			}
		}
	}
	for _, v := range i.IsVuln {
		if v.CVE != nil {
			if _, ok := cveMap[v.CVE.CveId]; !ok {
				cveMap[v.CVE.CveId] = v.CVE
			}
		}
	}
	for _, v := range i.Vex {
		if v.CVE != nil {
			if _, ok := cveMap[v.CVE.CveId]; !ok {
				cveMap[v.CVE.CveId] = v.CVE
			}
		}
	}
	cves := make([]*generated.CVEInputSpec, 0, len(cveMap))

	for _, cve := range cveMap {
		cves = append(cves, cve)
	}
	return cves
}

func (i IngestPredicates) GetOSVs(ctx context.Context) []*generated.OSVInputSpec {
	osvMap := make(map[string]*generated.OSVInputSpec)
	for _, vuln := range i.CertifyVuln {
		if vuln.OSV != nil {
			if _, ok := osvMap[vuln.OSV.OsvId]; !ok {
				osvMap[vuln.OSV.OsvId] = vuln.OSV
			}
		}
	}
	for _, v := range i.IsVuln {
		if v.OSV != nil {
			if _, ok := osvMap[v.OSV.OsvId]; !ok {
				osvMap[v.OSV.OsvId] = v.OSV
			}
		}
	}
	for _, v := range i.Vex {
		if v.OSV != nil {
			if _, ok := osvMap[v.OSV.OsvId]; !ok {
				osvMap[v.OSV.OsvId] = v.OSV
			}
		}
	}
	osvs := make([]*generated.OSVInputSpec, 0, len(osvMap))

	for _, osv := range osvMap {
		osvs = append(osvs, osv)
	}
	return osvs
}

func (i IngestPredicates) GetGHSAs(ctx context.Context) []*generated.GHSAInputSpec {
	ghsaMap := make(map[string]*generated.GHSAInputSpec)
	for _, vuln := range i.CertifyVuln {
		if vuln.GHSA != nil {
			if _, ok := ghsaMap[vuln.GHSA.GhsaId]; !ok {
				ghsaMap[vuln.GHSA.GhsaId] = vuln.GHSA
			}
		}
	}
	for _, v := range i.IsVuln {
		if v.GHSA != nil {
			if _, ok := ghsaMap[v.GHSA.GhsaId]; !ok {
				ghsaMap[v.GHSA.GhsaId] = v.GHSA
			}
		}
	}
	for _, v := range i.Vex {
		if v.GHSA != nil {
			if _, ok := ghsaMap[v.GHSA.GhsaId]; !ok {
				ghsaMap[v.GHSA.GhsaId] = v.GHSA
			}
		}
	}
	ghsas := make([]*generated.GHSAInputSpec, 0, len(ghsaMap))

	for _, ghsa := range ghsaMap {
		ghsas = append(ghsas, ghsa)
	}
	return ghsas
}

func concatenateSourceInput(source *generated.SourceInputSpec) string {
	var sourceElements []string
	sourceElements = append(sourceElements, source.Type, source.Namespace, source.Name)
	if source.Tag != nil {
		sourceElements = append(sourceElements, *source.Tag)
	}
	if source.Commit != nil {
		sourceElements = append(sourceElements, *source.Commit)
	}
	return strings.Join(sourceElements, "/")
}

// AssemblerInput represents the inputs to add to the graph
type AssemblerInput = IngestPredicates
