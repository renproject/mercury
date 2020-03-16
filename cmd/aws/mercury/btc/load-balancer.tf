// Create a security group for the bitcoin load balancer.
resource "aws_security_group" "aws_sg_lb_bitcoin" {
  name = "aws_sg_lb_bitcoin"
  vpc_id = var.vpc_id
  description = "Security group for bitcoin load balancer"

  ingress {
    description = "Allow http request"
    from_port = 80
    to_port = 80
    protocol = "tcp"
    cidr_blocks = [
      "0.0.0.0/0"]
  }
}

// Creata load balancer for the bitcoin mainne nodes.
resource "aws_lb" "aws_lb_bitcoin" {
  name = "aws-lb-bitcoin"
  internal = true
  load_balancer_type = "application"
  security_groups = [
    var.default_sg_id,
    aws_security_group.aws_sg_lb_bitcoin.id]
  subnets = [
    var.subnet_id_1,
    var.subnet_id_2]

  tags = {
    project = "mercury"
  }
}

// Load balancer target group.
resource "aws_lb_target_group" "aws_lb_tg_bitcoin" {
  name = "bitcoin-lb-tg-bitcoin"
  port = 80
  protocol = "HTTP"
  vpc_id = var.vpc_id
}

// Attach instances to the target group.
resource "aws_lb_target_group_attachment" "aws_lb_tga_bitcoin1" {
  target_group_arn = aws_lb_target_group.aws_lb_tg_bitcoin.arn
  target_id = aws_instance.bitcoin-mainnet-1.id
  port = 8332
}

resource "aws_lb_target_group_attachment" "aws_lb_tga_bitcoin2" {
  target_group_arn = aws_lb_target_group.aws_lb_tg_bitcoin.arn
  target_id = aws_instance.bitcoin-mainnet-2.id
  port = 8332
}

// Provide a listener to the load balancer
resource "aws_lb_listener" "aws_lb_listener_bitcoin" {
  load_balancer_arn = aws_lb.aws_lb_bitcoin.arn
  port = "80"
  protocol = "HTTP"

  default_action {
    type = "forward"
    target_group_arn = aws_lb_target_group.aws_lb_tg_bitcoin.arn
  }
}

// Output the DNS name of the bitcoin load balancer.
output "btc_lb_dns" {
  value = aws_lb.aws_lb_bitcoin.dns_name
}
