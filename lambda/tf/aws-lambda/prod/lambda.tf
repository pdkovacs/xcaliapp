locals {
  function_name    = "xcalidrawing"
  archive_filename = "${local.function_name}.zip"
  impl_directory   = "${path.module}/../../../prod"
}

data "aws_s3_bucket" "store" {
  bucket = var.s3_bucket
}

resource "aws_iam_policy" "read_write_s3_bucket" {
  name = "xcaliapp_read_write_s3_bucket"
  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Sid" : "s3listbucket",
        "Effect" : "Allow",
        "Action" : [
          "s3:ListBucket",
          "s3:ListBucketVersions",
        ],
        "Resource" : [
          "arn:aws:s3:::${var.s3_bucket}"
        ]
      },
      {
        "Sid" : "s3readwritebucketobjects",
        "Effect" : "Allow",
        "Action" : [
          "s3:GetObject",
          "s3:GetObjectVersion",
          "s3:PutObject",
          "s3:DeleteObject",
        ],
        "Resource" : [
          "arn:aws:s3:::${var.s3_bucket}/*"
        ]
      },
      {
        "Sid" : "s3denyhardandversioningchanges",
        "Effect" : "Deny",
        "Action" : [
          "s3:DeleteObjectVersion",
          "s3:PutBucketVersioning",
          "s3:PutBucketLifecycleConfiguration",
          "s3:DeleteBucket",
        ],
        "Resource" : [
          "arn:aws:s3:::${var.s3_bucket}",
          "arn:aws:s3:::${var.s3_bucket}/*",
        ]
      }
    ]
  })
}

resource "aws_iam_role" "iam_for_lambda" {
  name = "iam_for_xcaliapp_lambda"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "basic_lambda_exec" {
  role       = aws_iam_role.iam_for_lambda.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy_attachment" "read_write_s3_bucket" {
  role       = aws_iam_role.iam_for_lambda.name
  policy_arn = aws_iam_policy.read_write_s3_bucket.arn
}

data "archive_file" "lambda" {
  type        = "zip"
  source_file = "${local.impl_directory}/bootstrap"
  output_path = "${local.impl_directory}/${local.archive_filename}"
}

resource "aws_lambda_function" "xcali_prod" {
  function_name = local.function_name
  architectures = ["arm64"]
  filename      = "${local.impl_directory}/${local.archive_filename}"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "index.handler"

  source_code_hash = data.archive_file.lambda.output_base64sha256

  runtime = "provided.al2023"

  environment {
    variables = {
      DRAWINGS_BUCKET_NAME  = var.s3_bucket
      CF_ACCESS_TEAM_DOMAIN = var.cf_access_team_domain
      CF_ACCESS_AUD         = cloudflare_zero_trust_access_application.app.aud
    }
  }

  depends_on = [
    aws_iam_role_policy_attachment.lambda_logs,
    # aws_cloudwatch_log_group.xcali_prod
  ]
}

resource "aws_cloudwatch_log_group" "xcali_prod" {
  name              = "/aws/lambda/${local.function_name}"
  retention_in_days = 14
}

data "aws_iam_policy_document" "lambda_logging" {
  statement {
    effect = "Allow"

    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    resources = ["arn:aws:logs:*:*:*"]
  }
}

resource "aws_iam_policy" "lambda_logging" {
  name        = "xcaliapp_lambda_logging"
  path        = "/"
  description = "IAM policy for logging from a lambda"
  policy      = data.aws_iam_policy_document.lambda_logging.json
}

resource "aws_iam_role_policy_attachment" "lambda_logs" {
  role       = aws_iam_role.iam_for_lambda.name
  policy_arn = aws_iam_policy.lambda_logging.arn
}

resource "aws_lambda_function_url" "xcali_prod" {
  function_name      = aws_lambda_function.xcali_prod.function_name
  authorization_type = "NONE"
}

resource "aws_lambda_permission" "function_url_public" {
  statement_id           = "AllowPublicFunctionUrlAccess"
  action                 = "lambda:InvokeFunctionUrl"
  function_name          = aws_lambda_function.xcali_prod.function_name
  principal              = "*"
  function_url_auth_type = "NONE"
}

resource "aws_lambda_permission" "function_invoke_public" {
  statement_id  = "AllowPublicInvokeFunction"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.xcali_prod.function_name
  principal     = "*"
}

