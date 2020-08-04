//////////////////////////////////////////

	Description

/////////////////////////////////////////

	This is a simple project which reperesent an sample of 
intercommunicate between two different processes by 
ProtoBuf protocol.

	Project contains two separated parts: Transmitter,
which generates a set of Fibonacci numbers in the loop with preset
frequency, and Receiver, which receives them with it's own speed.

	All received values should be logged into the file, unless
in case Transmitter frequency exceeds Receiver speed (ability to log)
received values are ignored.
/////////////////////////////////////////

To proper execution firstly Receiver should be run,
then Transmitter can be started.

Both applications are finished after maximum Int64 number is achieved.


