// Create a security group for the zcash load balancer.
resource "aws_security_group" "aws_sg_lb_zcash" {
  name = "aws_sg_lb_zcash"
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
}

// Creata load balancer for the zcash mainne nodes.
resource "aws_lb" "aws_lb_zcash" {
  name = "aws-lb-zcash"
  internal = true
  load_balancer_type = "application"
  security_groups = [
    var.default_sg_id,
    aws_security_group.aws_sg_lb_zcash.id]
  subnets = [
    var.subnet_id_1,
    var.subnet_id_2]

  tags = {
    project = "mercury"
  }
}

// Load balancer target group.
resource "aws_lb_target_group" "aws_lb_tg_zcash" {
  name = "aws-lb-tg-zcash"
  port = 80
  protocol = "HTTP"
  vpc_id = var.vpc_id
}

// Attach instances to the target group.
resource "aws_lb_target_group_attachment" "aws_lb_tga_zcash1" {
  target_group_arn = aws_lb_target_group.aws_lb_tg_zcash.arn
  target_id = aws_instance.zcash-mainnet-1.id
  port = 8232
}

resource "aws_lb_target_group_attachment" "aws_lb_tga_zcash2" {
  target_group_arn = aws_lb_target_group.aws_lb_tg_zcash.arn
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
    target_group_arn = aws_lb_target_group.aws_lb_tg_zcash.arn
  }
}

// Output the DNS name of the zcash load balancer.
output "zec_lb_dns" {
  value = aws_lb.aws_lb_zcash.dns_name
}