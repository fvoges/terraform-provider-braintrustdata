# Read a tag by ID
data "braintrustdata_tag" "by_id" {
  id = "tag-123"
}

# Read a tag by name with optional project filters
data "braintrustdata_tag" "by_name" {
  name       = "production"
  project_id = "proj-123"
}

output "tag_id" {
  value = data.braintrustdata_tag.by_name.id
}

output "tag_color" {
  value = data.braintrustdata_tag.by_name.color
}
