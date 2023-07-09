<!-- This README file is going to be the one displayed on the Grafana.com website for your plugin -->

# Destiny Data source

Fetches activity history for players from Destiny 2 using the Bungie.net API.

![Screenshot of the plugin](https://raw.githubusercontent.com/joshhunt/destiny-grafana-datasource-plugin/main/src/img/screenshot.png)

- Lists activity history for single player.
- Can return history for all characters, or specific characters
- Can filter activity history by activity mode

Data returned for each activity:

- Character played
- Activity details (activity name, PvP map, and 'directory activity')
- Specific activity (e.g. map for PvP)
- Start and end times, and activity duration
- Time character was in the activity (e.g. if joined in progress)
- PGCR ID
- Whether activity was successfully completed or not
- Whether character won or lost

## Setup

This data source plugin requires a developer API key from Bungie.net:

- Log into bungie.net
- Visit [www.bungie.net/en/Application/Create](https://www.bungie.net/en/Application/Create)
- Fill in the fields. It's not particuarly important for most of the values:
  - Application Name: a familiar name for you to refer to this API key
  - Website: Doesn't matter - example.com is fine
  - OAuth Client Type: "Not applicable"
  - Redirect URL: leave blank
  - Origin Header: leave blank
- Once created, the page will list the API Key. Copy and use this when configuring this Data source in Grafana.
