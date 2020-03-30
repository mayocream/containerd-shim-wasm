#include <stdio.h>
#include <stdlib.h>
#include <errno.h>
#include <sys/utsname.h>

extern char **environ;

int main(int argc, char *argv[]) {

   struct utsname buffer;

   errno = 0;
   if (uname(&buffer) != 0) {
      perror("uname");
      exit(EXIT_FAILURE);
   }

   // Print system information
   printf("sysname = %s\n", buffer.sysname);
   printf("machine = %s\n", buffer.machine);

   // Print arguments
   for(int counter=0; counter<argc; counter++) {
		printf("argv[%d]: %s\n",counter,argv[counter]);
   }

   // Print environment
   printf("environment:\n");
   char *s = *environ;
   for (int i = 1; s; i++) {
      printf("%s\n", s);
      s = *(environ+i);
   }

   int count = 0;
   while(1) {
      count++;
      sleep(1); // set sleep value as u wish
      printf("sysname = %s %d\n", buffer.sysname, count);
   }

   return EXIT_SUCCESS;
}