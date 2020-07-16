// Create a new security group for the load balancer.
resource "aws_security_group" "aws_sg_lb_mercury" {
  name = "aws_security_group_lb_mercury"
  description = "Security group for mercury load balancer"
  vpc_id = aws_vpc.aws_vpc_mercury.id

  ingress {
    description = "Allow public connection from http"
    from_port = 80
    to_port = 80
    protocol = "tcp"
    cidr_blocks = [
      "0.0.0.0/0"]
  }
}

// Create a load-balaner in front of all mercury instances.
resource "aws_lb" "aws_lb_mercury" {
  name = "aws-lb-mercury"
  internal = false
  load_balancer_type = "application"
  security_groups = [
    aws_security_group.aws_sg_mercury_default.id,
    aws_security_group.aws_sg_lb_mercury.id
  ]

  subnets = [
    aws_subnet.aws_subnet1.id,
    aws_subnet.aws_subnet2.id]

  tags = {
    project = "mercury"
  }
}

// Load balancer target group.
resource "aws_lb_target_group" "aws_lb_tg_mercury" {
  name = "aws-lb-tg-mercury"
  port = 80
  protocol = "HTTP"
  vpc_id = aws_vpc.aws_vpc_mercury.id

  health_check {
    path = "/health"
  }
}

// Attach instances to the load balancer
resource "aws_lb_target_group_attachment" "aws_lb_tga_1" {
  target_group_arn = aws_lb_target_group.aws_lb_tg_mercury.arn
  target_id = aws_instance.aws_instance_mercury1.id
  port = 5000
}

resource "aws_lb_target_group_attachment" "aws_lb_tga_2" {
  target_group_arn = aws_lb_target_group.aws_lb_tg_mercury.arn
  target_id = aws_instance.aws_instance_mercury2.id
  port = 5000
}

// Provide a listener to the load balancer
resource "aws_lb_listener" "aws_lb_listener_mercury" {
  load_balancer_arn = aws_lb.aws_lb_mercury.arn
  port = "80"
  protocol = "HTTP"

  default_action {
    type = "forward"
    target_group_arn = aws_lb_target_group.aws_lb_tg_mercury.arn
  }
}

// Output the dns of the load-balancer
output "mercury_load_balancer_dns" {
  value = aws_lb.aws_lb_mercury.dns_name
}