resource "braintrustdata_project" "evaluation" {
  name        = "tag-example-project"
  description = "Project for tag resource examples"
}

resource "braintrustdata_tag" "priority" {
  project_id  = braintrustdata_project.evaluation.id
  name        = "priority"
  description = "Highlights high-priority traces and experiments"
  color       = "#ff0000"
}

output "tag_id" {
  value = braintrustdata_tag.priority.id
}
