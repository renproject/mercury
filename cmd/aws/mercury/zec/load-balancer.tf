resource "aws_lb" "aws_lb_zcash" {
  name = "aws-lb-zcash"
  internal = true
  load_balancer_type = "application"
  security_groups = [
    aws_security_group.aws_sg_lb_zcash.id]
  subnets = [
    var.subnet_id_1,
    var.subnet_id_2]

  tags = {
    Name = "mercury"
  }
}

// Load balancer target group.
resource "aws_lb_target_group" "aws_lb_zcash_target_group" {
  name = "zcash-lb-target-group"
  port = 80
  protocol = "HTTP"
  vpc_id = var.vpc_id
}

// Attach instances to the load balancer
resource "aws_lb_target_group_attachment" "aws_lb_zcash_target_group_attachment1" {
  target_group_arn = aws_lb_target_group.aws_lb_zcash_target_group.arn
  target_id = aws_instance.zcash-mainnet-1.id
  port = 8232
}

resource "aws_lb_target_group_attachment" "aws_lb_zcash_target_group_attachment2" {
  target_group_arn = aws_lb_target_group.aws_lb_zcash_target_group.arn
  target_id = aws_instance.zcash-mainnet-2.id
  port = 8232
}

// Provide a listener to the load balancer
resource "aws_lb_listener" "aws_lb_listener_zcash" {
  load_balancer_arn = aws_lb.aws_lb_zcash.arn
  port = "80"
  protocol = "HTTP"

  default_action {
    type = "forward"
    target_group_arn = aws_lb_target_group.aws_lb_zcash_target_group.arn
  }
}

// Create a default security group for the new VPC
resource "aws_security_group" "aws_sg_lb_zcash" {
  name = "aws_sg_lb_zcash_"
  vpc_id = var.vpc_id
  description = "Security group for zcash load balancer"

  ingress {
    description = "Allow http request"
    from_port = 80
    to_port = 80
    protocol = "tcp"
    cidr_blocks = [
      "0.0.0.0/0"]
  }

  egress {
    from_port = 0
    to_port = 0
    protocol = "-1"
    cidr_blocks = [
      "0.0.0.0/0"]
  }
}

output "zec_lb_dns" {
  value = aws_lb.aws_lb_zcash.dns_name
}