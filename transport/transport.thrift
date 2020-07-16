namespace go transportService



// Data structure for making transfer Transmitter - Receiver
struct CarrierDto
{
	10: i64 fibNumber,
}


service sender
{
	void Send(1: CarrierDto message)
}
