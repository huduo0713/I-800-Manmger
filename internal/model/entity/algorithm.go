// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// Algorithm is the golang structure for table algorithm.
type Algorithm struct {
	Id                 int    `json:"id"                 orm:"id"                   description:""` //
	AlgorithmId        string `json:"algorithmId"        orm:"algorithm_id"         description:""` //
	AlgorithmName      string `json:"algorithmName"      orm:"algorithm_name"       description:""` //
	AlgorithmVersion   string `json:"algorithmVersion"   orm:"algorithm_version"    description:""` //
	AlgorithmVersionId string `json:"algorithmVersionId" orm:"algorithm_version_id" description:""` //
	AlgorithmDataUrl   string `json:"algorithmDataUrl"   orm:"algorithm_data_url"   description:""` //
	FileSize           int    `json:"fileSize"           orm:"file_size"            description:""` //
	Md5                string `json:"md5"                orm:"md5"                  description:""` //
	LocalPath          string `json:"localPath"          orm:"local_path"           description:""` //
}
