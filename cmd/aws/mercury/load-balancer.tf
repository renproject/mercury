resource "aws_security_group" "aws_security_group_lb_mercury" {
  name        = "aws_security_group_lb_mercury"
  description = "Security group for mercury load balancer"
  vpc_id      = aws_vpc.aws_vpc_mercury.id

  ingress {
    description = "Allow SSH connection "
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "Allow public connection from http"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_lb" "aws_lb_mercury" {
  name               = "aws-lb-mercury"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.aws_security_group_lb_mercury.id]
  subnets            = [aws_subnet.aws_subnet1.id, aws_subnet.aws_subnet2.id]

  tags = {
    Name = "mercury"
  }
}

// Load balancer target group.
resource "aws_lb_target_group" "aws_lb_mercury_target_group" {
  name     = "mercury-lb-target-group"
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.aws_vpc_mercury.id

  health_check {
    path = "/health"
  }
}

// Attach instances to the load balancer
resource "aws_lb_target_group_attachment" "aws_lb_mercury_target_group_attachment1" {
  target_group_arn  = aws_lb_target_group.aws_lb_mercury_target_group.arn
  target_id         = aws_instance.aws_instance_mercury1.id
  port              = 5000
}

resource "aws_lb_target_group_attachment" "aws_lb_mercury_target_group_attachment2" {
  target_group_arn  = aws_lb_target_group.aws_lb_mercury_target_group.arn
  target_id         = aws_instance.aws_instance_mercury2.id
  port              = 5000
}

// Provide a listener to the load balancer
resource "aws_lb_listener" "aws_lb_listener_mercury" {
  load_balancer_arn = aws_lb.aws_lb_mercury.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.aws_lb_mercury_target_group.arn
  }
}