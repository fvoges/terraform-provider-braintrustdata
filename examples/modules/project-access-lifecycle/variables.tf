variable "project_name_prefix" {
  description = "Optional prefix added to every created project name."
  type        = string
  default     = ""
}

variable "role_group_name_prefix" {
  description = "Optional prefix added to every created project role-group name."
  type        = string
  default     = ""
}

variable "binding_key_delimiter" {
  description = "Delimiter used in binding keys."
  type        = string
  default     = "|"

  validation {
    condition     = trimspace(var.binding_key_delimiter) != ""
    error_message = "binding_key_delimiter must not be empty."
  }
}

variable "projects" {
  description = "Project catalog keyed by stable project key."
  type = map(object({
    name          = optional(string)
    description   = optional(string)
    enabled_roles = optional(set(string))
    role_overrides = optional(map(object({
      permissions          = optional(set(string))
      restrict_object_type = optional(string)
      description          = optional(string)
    })), {})
  }))
  default = {}

  validation {
    condition = alltrue(flatten([
      for _, project in var.projects : [
        for _, override in try(project.role_overrides, {}) : (
          override.permissions == null || alltrue([
            for permission in override.permissions : contains([
              "create",
              "read",
              "update",
              "delete",
              "create_acls",
              "read_acls",
              "update_acls",
              "delete_acls",
            ], permission)
          ])
        )
      ]
    ]))
    error_message = "projects[*].role_overrides[*].permissions must only contain: create, read, update, delete, create_acls, read_acls, update_acls, delete_acls."
  }

  validation {
    condition = alltrue(flatten([
      for _, project in var.projects : [
        for _, override in try(project.role_overrides, {}) : (
          override.restrict_object_type == null || contains([
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
          ], override.restrict_object_type)
        )
      ]
    ]))
    error_message = "projects[*].role_overrides[*].restrict_object_type must be null or a valid Braintrust ACL object type."
  }
}

variable "role_catalog" {
  description = "Central access role catalog."
  type = map(object({
    description          = optional(string)
    permissions          = set(string)
    restrict_object_type = optional(string)
  }))
  default = {}

  validation {
    condition = alltrue(flatten([
      for _, role in var.role_catalog : [
        for permission in role.permissions : contains([
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
    error_message = "role_catalog permissions must only contain: create, read, update, delete, create_acls, read_acls, update_acls, delete_acls."
  }

  validation {
    condition = alltrue([
      for _, role in var.role_catalog : (
        role.restrict_object_type == null || contains([
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
        ], role.restrict_object_type)
      )
    ])
    error_message = "role_catalog restrict_object_type must be null or a valid Braintrust ACL object type."
  }
}

variable "role_group_member_group_ids_by_binding_key" {
  description = "Map of binding key project_key|role_key to member group IDs added to the role group."
  type        = map(set(string))
  default     = {}
}
