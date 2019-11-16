<template>
  <el-table :data="tableData" stripe style="width: 100%">
    <el-table-column prop="name" label="仓库"></el-table-column>
    <el-table-column fixed="right" label="操作" width="100">
      <template slot-scope="scope">
        <el-button @click="handleClick(scope.row)" type="text" size="small">查看</el-button>
        <el-button type="text" size="small">编辑</el-button>
      </template>
    </el-table-column>
  </el-table>
</template>

<script>
export default {
  name: "Repository",

  methods: {
    handleClick(row) {
      // eslint-disable-next-line no-console
      console.log(row.name);
    }
  },

  data() {
    return {
      tableData: [
        {
          name: "coredns"
        },
        {
          name: "test/abc"
        }
      ]
    };
  }
};
</script>
