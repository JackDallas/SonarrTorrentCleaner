<script lang="ts">
  import type { SonnarQueueResponse } from "./SonnarQueue";
  import SvelteTable from "svelte-table";
  import { DateTime } from "luxon";


  let COLUMNS = [
    {
      key: "id",
      title: "ID",
      value: (v) => v.id,
    },
    {
      key: "title",
      title: "Title",
      value: (v) => v.title,
      sortable: true,
      headerClass: "text-left",
    },
    {
      key: "status",
      title: "Status",
      value: (v) => v.status,
      sortable: true,
      headerClass: "text-left",
    },
    {
      key: "downloaded",
      title: "Downloaded (%)",
      value: (v) => v.downloaded,
      sortable: true,
      headerClass: "text-left",
    },
    {
      key: "lastchecked",
      title: "Last Checked",
      value: (v) => v.lastChecked,
      sortable: true,
      headerClass: "text-left",
    },
    {
      key: "activefor",
      title: "Active For",
      value: (v) => v.activefor,
      sortable: true,
      headerClass: "text-left",
    },
  ];
  let rows = [];

  let updateData = () => {
    console.log("Updating Data")
    fetch(new Request("/api/queue"))
      .then((response) => response.json())
      .then((data) => {
        let queue: SonnarQueueResponse[] = data;
        queue.forEach((item) => {
          let start = DateTime.fromISO(item.FirstSeen);
          let end = DateTime.now();
          var diff = end.diff(start,['minutes', 'hours', 'days'])
          diff = diff.toObject();

          let activeFor = diff['minutes'].toFixed(0) + "m ";
          if (diff['hours'] !== 0)
            activeFor + diff['hours'] + "h "
          if (diff['days'] !== 0)
            activeFor + diff['days'] + "days"

          let newRow = {
            id: item.Item.downloadId,
            title: item.Item.title,
            status: item.Item.status,
            downloaded: Number(((item.Item.sizeleft / item.Item.size) * 100).toFixed(2)),
            lastChecked: DateTime.fromISO(item.LastChecked).toLocaleString(DateTime.DATETIME_SHORT),
            activefor: activeFor,
          };
          //.push is bugged in svelte 
          rows = [...rows, newRow];
        });
      });
  };
  updateData();

  let intID = setInterval(updateData, 30000);
</script>
<main>
  <h1>Sonnar Torrent Queue</h1>
  <SvelteTable columns={COLUMNS} rows={rows} />
</main>
