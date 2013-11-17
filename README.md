### Part Time Geoffrey 

Really simple CI server in golang. 

It runs scripts in directories where the binary is located. It outputs logs. It lists those logs by date. It pipelines projects.

Written because Jenkins is too resource hungry to run on this shitty laptop / server.

Only use if you can read and understand the source, all 200 lines of it.

Don't run it these situations

0. Your server is open to the internet.
0. You're on acid.
0. You have more than glib attitude to security, robustness, documentation and sanity.

And understand I wrote this in a few hours and it may well ruin everything you hold dear.

It's also got a security hole, but I think you have to get jazzy with some funky encodings in the URL to take advantage of it. 

I could fix it. I really could.
