package main

import "time"

/*

// For Display Purposes
// We must be able to display the available Ground Stations, Antennas, and DataPaths
GET /gs/{id}
GET /gs/{id}/antenna
GET /antenna/{id}/dataPaths
GET /antenna/{id}/dataPaths/{id}

// For Contact Manager
// We must be able to allocate a Ground Station for a Contact
// We must be able to determine the type of Antenna that was allocated
// We must be able to know how many uplink jobs need to be run
POST /contact -> allocation link
GET /contacts/{id}/allocations/datapaths?purpose=data
GET /contacts/{id}/allocations/datapaths?purpose=telemetry
GET /contacts/{id}/allocations/datapaths?purpose=command
GET /contacts/{id}/allocations/antenna // Need to know the type of Antenna being used

// For Contact Manager
// We must be able to allocate a Ground Station for Tests
POST /test -> allocation link
GET /test/{id}/allocations

// For Antenna Work Items Manager
// We must be able to determine which Antenna to use
// We must be able to determine if the Antenna supports Redundancy -> More than 2 datapaths
// DAF must determine which DTAS to also schedule an archive activity.
GET /contact/{id}/allocations/antenna // determine Antenna to use
GET /contact/{id}/allocations/datapaths // determine redundancy
GET /antennas/{id}/dataPaths?purpose=data&isPrimary=true // determine primary data path
GET /contact/{id}/allocations/datapaths?purpose=data // if first is not primary, schedule secondary

DAF Standard Antenna System

In this scenario each data path type has a single path. Imagery data will be reconstructed on the
Primary X-Band-Down data path, commands will be sent to the Primary S-Band-Up data path, and telemtry
will be reconstructed on the Primary S-Band-Down data path.

						          Ground Terminal
						                 |
		  ---------------------------- Antenna -----------------------
	      |					             |                           |
      X-Band-Down                   S-Band-Down                  S-Band-Up
          |                              |                           |
       *Primary                       *Primary                    *Primary
	      |                              |                           |
         Recon                         Telem(1)                    Cmd(1)

DAF Redundant Antenna System - Primary Preferred

In this scenario each data path type has a primary and secondary path. Data coming
down can be reconstructed on the primary and/or secondary paths and commands going up can be sent on
the primary and/or secondary paths. Imagery data is available for reconstruction on both the Primary
and Secondary X-Band-Down data paths. Primary is preferred so Reconstruction will be executed
on the Primary X-Band-Down data path. Commands will be sent on both Primary and Secondary S-Band-Up
datapaths, however the Antenna System will block commands on the Secondary data path. The Antenna Sytem
duplicates S-Band-Down data by routing the data received on the Primary data path to the Secondary
datapath. Telemetry data is reconstructed on both data paths. The Antenna System is capable of
automatically switching to the Secondary S-Band-Up datapath if for some reason the Primary fails.


						          Ground Terminal
						                 |
		  ---------------------------- Antenna -----------------------
	      |					             |                           |
    -- X-Band-Down --             -- S-Band-Down --             -- S-Band-Up --
    |               |             |               X             |             X
*Primary         Secondary    *Primary ------> Secondary     *Primary      Secondary
	|               |             |               |             |             |
  Recon            [X]         Telem(1)         Telem(2)       Cmd(1)       Cmd(2)


DAF Redundant Antenna System - Secondary Preferred

In this scenario each data path type has a primary and secondary path. Data coming
down can be reconstructed on the primary and/or secondary paths and commands going up can be sent on
the primary and/or secondary paths. Imagery data is available for reconstruction on both the Primary
and Secondary X-Band-Down data paths. Secondary is preferred so Reconstruction will be executed
on the Primary X-Band-Down data path. Commands will be sent on both Primary and Secondary S-Band-Up
datapaths, however the Antenna System will block commands on the Primary data path. The Antenna Sytem
duplicates S-Band-Down data by routing the data received on the Secondary data path to the Primary
datapath. Telemetry data is reconstructed on both data paths. The Antenna System is capable of
automatically switching to the Primary S-Band-Up datapath if for some reason the Secondary fails.

						           Ground Terminal
						                 |
		  ---------------------------- Antenna -----------------------
	      |					             |                           |
    -- X-Band-Down --             -- S-Band-Down --             -- S-Band-Up --
    |               |             X               |             X             |
 Primary         *Secondary    Primary <------ *Secondary     Primary      *Secondary
	|               |             |               |             |             |
   [x]            Recon         Task(1)         Task(2)       Task(1)       Task(2)


*/

// GroundStation will contain a single Antenna
type GroundStation struct {
	ID       string
	Name     string
	Location string
	Antenna  Antenna
}

// Antenna can contain any number of DataPaths
type Antenna struct {
	ID              string
	Name            string
	Type            string // Zodiac, Viasat, Viasat-MASS, etc
	GroundStationID string
	DataPaths       []DataPath
}

// DataPath describes the type and direction of data through an Antenna
type DataPath struct {
	ID        string
	AntennaID string // DataPath belongs to an Antennna
	Name      string // Name for display purposes
	Band      string // X or S (Frequency Bands)
	Direction string // Up (Ground to Satellite) or Down (Satellite to Ground)
	Purpose   string // Data (Imagery), Telemetry (Health and Commanding Status Info), Command (Satellite Commands)
	IsPrimary bool
}

// Contact describes the Communications between a Satellite and Ground Station
type Contact struct {
	ContactID      string
	SiteID         string
	SatelliteID    string
	Communications []Communication
}

// Communication describes a specific type of interaction between a Satellite and Ground Station
type Communication struct {
	Band          string // S or X
	Direction     string // Up or Down
	Purpose       string // Data, Telemetry, Command
	WithTest      bool   // True if test should be run before communication, false otherwise
	StartTime     time.Time
	EndTime       time.Time
	TestStartTime time.Time // optional, if not specified default is used
	TestEndTime   time.Time // optional, if not specified default is used
}

// ContactAllocation describes the allocation of a Ground Station for a Contact
type ContactAllocation struct {
	ID                  string
	ContactID           string
	GroundStationID     string
	AntennaID           string
	AntennaAllocation   AntennaAllocation
	DataPathAllocations []DataPathAllocation
}

// AntennaAllocation describes the allocation of a Antenna for a Contact
type AntennaAllocation struct {
	ID                  string
	ContactID           string
	StartTime           time.Time
	EndTime             time.Time
	DataPathAllocations []DataPathAllocation
}

// DataPathAllocation describes the allocation of a DataPath for an Antenna
type DataPathAllocation struct {
	ID         string
	ContactID  string
	DataPathID string
	StartTime  time.Time
	EndTime    time.Time
}
