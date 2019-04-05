# StreamerEventViewer 

The full spec for this project can be seen here: https://gist.github.com/osamakhn/14a378f3107d49de47e0b617a3d5fdf5

The current solution only displays follow events for the streamer. Other events can be added easily using the same subscription pattern that is already in use. Due to time constraints, I have chosen to limit what is added in this version. 

### Deploying to AWS ###
As a go program, this can be easily deployed to the Elastic BeanStalk, by following the method in the article: https://docs.aws.amazon.com/elasticbeanstalk/latest/dg/go-environment.html

### Potential Improvements ###
The current code was meant more as a proof of concept, and is not a full production-ready setup. In order to get it production ready, the configuration will have to be externalised, either to a configuration server, or even just an external file, in order to make changes easier, and in order to ensure better security

### Scaling ###
The current solution should support several hundred requests as it is currently implemented. In order to scale it up, I would break it down into sepearte services, which will allow us to scael each microservice independently of each other. The event listeners will have a separate microservice, running on different instances. In front of the main webapp, we will need to place a load balancer, in order to ensure that the load is optimally distributed. In order to enable and support the scaling, we will need to add metrics, dashboards and alarms to ensure visibility into the performance of the system. 

Since the amount of subscriptions that a single API account can have, I would recommend creating multiple accounts, and using them on a round robin basis, so that the users are always able to establish connections. 


