### chippy/chirpy
#### What?
chippy/chirpy - super nano tweet web server (built during `boot.dev`) which supports the below endpoints:

- `GET /api/chirps` - get all chirps in asc order of `chirpID`. Supports filter query params such as `authorID` and `sort`
- `GET /api/chirps/{chirpID}` - get a single chirp via the `chirpID`
- `DELETE /api/chirps/{chirpID}` - delete a chirp via the `chirpID`
- `POST /api/chirps` - create a chirp via specified request body
- `POST /api/login` - login as a user via user credentials
- `POST /polka/webhooks` - webhook to support "Polka" based payment service.

#### Why?
Built as part of `boot.dev`. The most useful takeaways from this project:

- Implementing a bare-bones JSON file based DB named Chibe (chee-bee). Safe of deadlocks.
- Implementing access and refresh tokens along with revoking mechanism, and authenticating via these tokens.
