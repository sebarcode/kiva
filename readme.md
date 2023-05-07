# KIVA
A 2-ways key-value helper
 
# Background
When we need to store hot data into memory store like local memory, memcache, redis, etcd, and other. Often we dealt with 2 data store. One for managing hot data and other one to taken care of the persistent storage (normally will be a db). This library is to ease the process of retrieve and store data on those 2 storages. 

# Concept
I am not really sure what is the name of this concept, but definitely is not new. 

## Retrieve data
- Check on hot data storage first, if exist take it
- If not read from persistent storage
- if not exist on persistent storage return error
- if exist on persistent storage, then save it to hots storage and return the value
- Mantain lifetime of every single data on hot storage, if not it will cost you memory usage
- Dispose data from hot storage if its lifetime has been reach

## Set Data
- Write to hot storage
- Mandate system to immidiately write to persistent storage or periodically thru batch

## Sync Data
- Sync will be run automatically on bckground, nothing we need to do
- Will be trigger as a batch periodically
- Get all keys
- Iterate for all keys
- If keys already expired, delete it
- If keys still valid, check for its syncDirection
- If SYNC_TO_PERSISTENT_STORAGE then run committer
- If SYNC_TO_HOTS_STORAGE then get latest and update hot storage

# The Catch
To use Kiva we will need 4 things:
- a GetterFunction implementation to read data directly from persistent storage
- a SetterFunction implmentation to store data into persistent storage
- a ItemReflectorFunction implementation, this function is used on sync process, to define template of new item for each table
- a Kiva Provider implementation to manage read and write data into hot storage

*NOTE: all data hosts on atable should consist data with datatype, if not, panic may happen. Need to work on this to anticipate panic