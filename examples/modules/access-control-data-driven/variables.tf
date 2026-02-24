variable "identity_case_insensitive" {
  description = "Normalize identity keys (for example emails) to lowercase before resolution."
  type        = bool
  default     = true
}

variable "project_name_prefix" {
  description = "Optional prefix added to every created project name."
  type        = string
  default     = ""
}

variable "group_name_prefix" {
  description = "Optional prefix added to every created group name."
  type        = string
  default     = ""
}

variable "projects" {
  description = "Project catalog keyed by stable project key."
  type = map(object({
    name        = optional(string)
    description = optional(string)
  }))
  default = {}
}

variable "groups" {
  description = "Role-like groups keyed by stable group key. member_identities are human-friendly keys (for example email/username) resolved through user_id_by_identity."
  type = map(object({
    name              = optional(string)
    description       = optional(string)
    member_identities = optional(set(string), [])
    member_user_ids   = optional(set(string), [])
    member_group_keys = optional(set(string), [])
  }))
  default = {}
}

variable "project_group_bindings" {
  description = "Access matrix from groups to projects. Each permission becomes one ACL entry."
  type = list(object({
    project_key          = string
    group_key            = string
    permissions          = set(string)
    restrict_object_type = optional(string)
  }))
  default = []

  validation {
    condition = alltrue(flatten([
      for binding in var.project_group_bindings : [
        for permission in binding.permissions : contains([
          "create",
          "read",
          "update",
          "delete",
          "create_acls",
          "read_acls",
          "update_acls",
          "delete_acls",
        ], permission)
      ]
    ]))
    error_message = "project_group_bindings.permissions must only contain: create, read, update, delete, create_acls, read_acls, update_acls, delete_acls."
  }

  validation {
    condition = alltrue([
      for binding in var.project_group_bindings : (
        binding.restrict_object_type == null || contains([
          "organization",
          "project",
          "experiment",
          "dataset",
          "prompt",
          "prompt_session",
          "group",
          "role",
          "org_member",
          "project_log",
          "org_project",
        ], binding.restrict_object_type)
      )
    ])
    error_message = "project_group_bindings.restrict_object_type must be null or a valid Braintrust ACL object type."
  }
}

variable "user_id_by_identity" {
  description = "Map of identity key -> Braintrust user ID. Identity key can be email/username based on your upstream resolver."
  type        = map(string)
  default     = {}
}
