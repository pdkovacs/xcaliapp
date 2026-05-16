locals {
  function_names    = ["getdrawing", "listdrawing", "putdrawing"]
  archive_filenames = [for f in local.function_names : "${f}.zip"]
  impl_directories  =  [for f in local.function_names : "${path.module}/../../../aws-lambda/${f}"]
}

output "archive_filenames" {
   value = local.archive_filenames
}
