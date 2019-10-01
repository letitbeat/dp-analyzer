# Data-plane Analyzer

Tool that receives network packets, filters the packets having the same UID (payload) and computes *flow trees* from the set of filtered packets.

## Algorithm's Intuition

**Input** Represented by triples *(s, i , t)* where *s ∈ S*, *i* is the interface at the node *s* where the packet is observed, and *t* is the timestamp when the observation occurred

**Output** Flow trees, which are a graphical representation of a set of data-paths from the same origin

1. Sort the observations(packets) by their timestamp
2. Create a tree with the root labeled with *h*
3. Keeping track of the time of packet ingress (TI) and egress (TE) properly "hang" the child nodes from the observations
4. Create a set of paths by doing a depth-first-search from the root to each of the leaves

> A detailed algorithm can be found in the [paper][1]

## Endpoints

### Topology related

Set the topology to be used during the flow trees generation.

**URL:** `/topology`

**Method:** `POST`

Returns the used topology in a Json form

**URL:** `/topology`

**Method:** `GET`

## Flow trees related

Used the store the packet's(observation) data

**URL:** `/save`

**Method:** `POST`

**Request Body Example:**

```json
{
    "device" :"s1-eth1",
    "type" :1,
    "src_ip" :"10.0.10.1",
    "dst_ip" :"10.0.10.2",
    "src_port" :"6666",
    "dst_port" :"80",
    "payload" :"2624c054-d068-4513-6631-71d824b428b4",
    "captured_at": "2019-03-16 17:43:26.385 +0000 UTC"
}
```

Returns all the generated flow trees in a Json form

**URL:** `/`

**Method:** `GET`

**Response Example:**

```json
{
    "id": "2624c054-d068-4513-6631-71d824b428b4",
    "type": "TCP",
    "src_ip": "10.0.0.7",
    "dst_ip": "10.0.0.252",
    "src_port": "0",
    "dst_port": "66(sql-net)",
    "nodes": "digraph root {\n\tgraph [label=\"TCP 66(sql-net)\", \n\t\tlabelloc=t\n\t];\n\tnode [label=\"\\N\"];\n\tsubgraph T {\n\t\tgraph [label=\"TCP 66(sql-net)\",\n\t\t\tlabelloc=t\n\t\t];\n\t..." // DOT representation
    "nodes_img": "iVBORw0KGgoAAAANSUhEUgAABc8AAANZCAYAAADZAcbeAAAABmJLR0QA/wD/AP+gvaeTAAAgAElEQVR4nOzdf3zP9f7/8ft7m9ls2hFrDEk0xrSa2WponZodlZPJz6Yj69Ox70EpLKJPH5JP..." // Base64 representation
}
```

### License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details

[1]: https://ieeexplore.ieee.org/xpl/conhome/1000490/all-proceedings
